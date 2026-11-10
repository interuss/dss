package consensus

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"os"
	"sync"
	"time"

	params "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/stacktrace"
	etcdtransport "go.etcd.io/etcd/client/pkg/v3/transport"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// NewTLSInfo returns the mutual TLS settings between Raft nodes, empty when mutual TLS is disabled.
// A peer must present a certificate signed by a CA of the CA file. The certificate is not matched against the
// peer's source IP, which differs from the addresses in the certificate behind NAT or a load balancer.
func NewTLSInfo(connectParams params.ConnectParameters, logger *zap.Logger) etcdtransport.TLSInfo {
	if !connectParams.TLSEnabled() {
		return etcdtransport.TLSInfo{}
	}
	return etcdtransport.TLSInfo{
		CertFile:            connectParams.TLSCertFile,
		KeyFile:             connectParams.TLSKeyFile,
		TrustedCAFile:       connectParams.TLSCAFile,
		ClientCertAuth:      true,
		SkipClientSANVerify: true,
		Logger:              logger,
	}
}

// newListener listens on addr, with mutual TLS unless tlsInfo is empty. The TLS handshake completes before a
// connection is handed to the Raft handler. Each connection rejected during the handshake is logged and counted.
func newListener(ctx context.Context, addr string, tlsInfo etcdtransport.TLSInfo, logger *zap.Logger) (net.Listener, error) {
	if tlsInfo.Empty() {
		return (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	}
	if err := checkCertificates(tlsInfo, logger); err != nil {
		return nil, err
	}

	failures, err := otel.Meter("raftstore").Int64Counter("raft_peer_tls_handshake_failures", metric.WithDescription("Connections from Raft peers rejected during the TLS handshake, by reason"))
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create the raft_peer_tls_handshake_failures metric")
	}
	// A peer retries many times per second, so each reason is logged at most once every 5 seconds, with the number of
	// failures removed since. The metric counts all.
	var mu sync.Mutex
	lastLogged := map[string]time.Time{}
	removed := map[string]int{}
	tlsInfo.HandshakeFailure = func(conn *tls.Conn, err error) {
		reason := handshakeFailureReason(err)
		failures.Add(ctx, 1, metric.WithAttributes(attribute.String("reason", reason)))
		mu.Lock()
		defer mu.Unlock()
		if time.Since(lastLogged[reason]) < 5*time.Second {
			removed[reason]++
			return
		}
		lastLogged[reason] = time.Now()
		fields := []zap.Field{zap.Int("duplicates_removed_for_"+reason, removed[reason]), zap.String("peer_address", conn.RemoteAddr().String()), zap.Error(err)}
		removed[reason] = 0
		if e, ok := errors.AsType[*tls.CertificateVerificationError](err); ok && len(e.UnverifiedCertificates) > 0 {
			peer := e.UnverifiedCertificates[0]
			fields = append(fields, zap.String("peer_subject", peer.Subject.String()), zap.String("peer_issuer", peer.Issuer.String()), zap.Time("peer_not_after", peer.NotAfter))
		}
		logger.Warn("raft: tls handshake failure: "+reason, fields...)
	}
	return etcdtransport.NewListenerWithOpts(addr, "https", etcdtransport.WithTLSInfo(&tlsInfo))
}

func handshakeFailureReason(err error) string {
	if _, ok := errors.AsType[tls.RecordHeaderError](err); ok {
		return "not_tls"
	}
	if errors.Is(err, io.EOF) {
		return "closed"
	}
	if err.Error() == "remote error: tls: bad certificate" {
		return "remote_rejected"
	}
	if _, ok := errors.AsType[x509.UnknownAuthorityError](err); ok {
		return "unknown_authority"
	}
	if e, ok := errors.AsType[x509.CertificateInvalidError](err); ok && e.Reason == x509.Expired {
		return "invalid_date"
	}
	if _, ok := errors.AsType[*tls.CertificateVerificationError](err); ok {
		return "bad_certificate"
	}
	return "other"
}

// checkCertificates fails when the certificate of this node cannot be loaded or has expired, or when no CA of the CA
// file is left that has not expired. It warns about each expired CA.
func checkCertificates(tlsInfo etcdtransport.TLSInfo, logger *zap.Logger) error {
	now := time.Now()
	cert, err := tls.LoadX509KeyPair(tlsInfo.CertFile, tlsInfo.KeyFile)
	if err != nil {
		return stacktrace.Propagate(err, "failed to load the certificate of this node from %s and %s", tlsInfo.CertFile, tlsInfo.KeyFile)
	}
	logger.Debug("raft: node certificate", zap.String("subject", cert.Leaf.Subject.String()), zap.String("issuer", cert.Leaf.Issuer.String()), zap.Time("not_after", cert.Leaf.NotAfter))
	if now.After(cert.Leaf.NotAfter) {
		return stacktrace.NewError("the certificate of this node in %s expired on %s", tlsInfo.CertFile, cert.Leaf.NotAfter)
	}

	cas, err := readCertificates(tlsInfo.TrustedCAFile)
	if err != nil {
		return stacktrace.Propagate(err, "failed to load the CAs from %s", tlsInfo.TrustedCAFile)
	}
	unexpired := 0
	for _, ca := range cas {
		fields := []zap.Field{zap.String("subject", ca.Subject.String()), zap.Time("not_after", ca.NotAfter)}
		if now.After(ca.NotAfter) {
			logger.Warn("raft: a CA of --raft_tls_ca_file has expired", fields...)
		} else {
			logger.Debug("raft: CA certificate", fields...)
			unexpired++
		}
	}
	if unexpired == 0 {
		return stacktrace.NewError("%s has no CA that has not expired", tlsInfo.TrustedCAFile)
	}
	return nil
}

// readCertificates reads the certificates of a PEM file
func readCertificates(file string) ([]*x509.Certificate, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var certs []*x509.Certificate
	for block, rest := pem.Decode(data); block != nil; block, rest = pem.Decode(rest) {
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}
		certs = append(certs, cert)
	}
	return certs, nil
}
