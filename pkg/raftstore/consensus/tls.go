package consensus

import (
	"context"
	"net"

	params "github.com/interuss/dss/pkg/raftstore/params"
	etcdtransport "go.etcd.io/etcd/client/pkg/v3/transport"
	"go.uber.org/zap"
)

// newTLSInfo returns the mutual TLS settings between Raft nodes, empty when mutual TLS is disabled.
// A peer must present a certificate signed by a CA of the CA file. The certificate is not matched against the
// peer's source IP, which differs from the addresses in the certificate behind NAT or a load balancer.
func newTLSInfo(connectParams params.ConnectParameters, logger *zap.Logger) etcdtransport.TLSInfo {
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
func newListener(ctx context.Context, addr string, tlsInfo etcdtransport.TLSInfo) (net.Listener, error) {
	if tlsInfo.Empty() {
		return (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	}
	return etcdtransport.NewListenerWithOpts(addr, "https", etcdtransport.WithTLSInfo(&tlsInfo))
}
