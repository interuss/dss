package consensus

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"time"

	params "github.com/interuss/dss/pkg/raftstore/params"
	"github.com/interuss/stacktrace"
	etcdtransport "go.etcd.io/etcd/client/pkg/v3/transport"
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
// connection is handed to the Raft handler.
func newListener(ctx context.Context, addr string, tlsInfo etcdtransport.TLSInfo, logger *zap.Logger) (net.Listener, error) {
	if tlsInfo.Empty() {
		return (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	}
	if err := checkCertificates(tlsInfo, logger); err != nil {
		return nil, err
	}
	return etcdtransport.NewListenerWithOpts(addr, "https", etcdtransport.WithTLSInfo(&tlsInfo))
}

// checkCertificates fails when the certificate of this node cannot be loaded or has expired, or when no CA of the CA
// file is left that has not expired. It warns about each expired CA.
func checkCertificates(tlsInfo etcdtransport.TLSInfo, logger *zap.Logger) error {
	now := time.Now()
	cert, err := tls.LoadX509KeyPair(tlsInfo.CertFile, tlsInfo.KeyFile)
	if err != nil {
		return stacktrace.Propagate(err, "failed to load the certificate of this node from %s and %s", tlsInfo.CertFile, tlsInfo.KeyFile)
	}
	if now.After(cert.Leaf.NotAfter) {
		return stacktrace.NewError("the certificate of this node in %s expired on %s", tlsInfo.CertFile, cert.Leaf.NotAfter)
	}

	cas, err := readCertificates(tlsInfo.TrustedCAFile)
	if err != nil {
		return stacktrace.Propagate(err, "failed to load the CAs from %s", tlsInfo.TrustedCAFile)
	}
	unexpired := 0
	for _, ca := range cas {
		if now.After(ca.NotAfter) {
			logger.Warn("a CA of --raft_tls_ca_file has expired", zap.String("subject", ca.Subject.String()), zap.Time("not_after", ca.NotAfter))
		} else {
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
