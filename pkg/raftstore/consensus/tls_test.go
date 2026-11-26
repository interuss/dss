package consensus

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/interuss/dss/pkg/raftstore/consensus/node"
	"github.com/interuss/dss/pkg/raftstore/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/client/pkg/v3/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func generateTestCA(t *testing.T, commonName string, serial int64, notBefore, notAfter time.Time) (*x509.Certificate, *ecdsa.PrivateKey, []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(serial),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)

	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)

	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return cert, key, pemBytes
}

func writeTestLeafCert(t *testing.T, dir, name string, caCert *x509.Certificate, caKey *ecdsa.PrivateKey, ips []net.IP, notBefore, notAfter time.Time) (string, string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(100),
		Subject:      pkix.Name{CommonName: name},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		IPAddresses:  ips,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &key.PublicKey, caKey)
	require.NoError(t, err)

	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)

	certPath := filepath.Join(dir, name+".crt")
	keyPath := filepath.Join(dir, name+".key")
	require.NoError(t, os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600))
	require.NoError(t, os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0o600))
	return certPath, keyPath
}

func TestNewTLSInfo(t *testing.T) {
	insecureInfo := NewTLSInfo(params.ConnectParameters{Insecure: true}, zap.NewNop())
	assert.True(t, insecureInfo.Empty())

	mtlsInfo := NewTLSInfo(params.ConnectParameters{
		TLSCertFile: "/path/node.crt",
		TLSKeyFile:  "/path/node.key",
		TLSCAFile:   "/path/ca.crt",
	}, zap.NewNop())
	assert.False(t, mtlsInfo.Empty())
	assert.Equal(t, "/path/node.crt", mtlsInfo.CertFile)
	assert.Equal(t, "/path/node.key", mtlsInfo.KeyFile)
	assert.Equal(t, "/path/ca.crt", mtlsInfo.TrustedCAFile)
	assert.True(t, mtlsInfo.ClientCertAuth)
	assert.True(t, mtlsInfo.SkipClientSANVerify)
}

func TestHandshakeFailureReason(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{name: "plaintext on TLS port", err: tls.RecordHeaderError{Msg: "first record does not look like a TLS handshake"}, want: "not_tls"},
		{name: "remote peer rejected our cert", err: errors.New("remote error: tls: bad certificate"), want: "remote_rejected"},
		{name: "untrusted CA", err: &tls.CertificateVerificationError{Err: x509.UnknownAuthorityError{}}, want: "unknown_authority"},
		{name: "expired peer cert", err: &tls.CertificateVerificationError{Err: x509.CertificateInvalidError{Reason: x509.Expired}}, want: "invalid_date"},
		{name: "other cert verification failure", err: &tls.CertificateVerificationError{Err: x509.CertificateInvalidError{Reason: x509.IncompatibleUsage}}, want: "bad_certificate"},
		{name: "unclassified error", err: errors.New("connection reset by peer"), want: "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, handshakeFailureReason(tt.err))
		})
	}
}

func TestCheckCertificates(t *testing.T) {
	now := time.Now()
	dir := t.TempDir()

	validCA, validCAKey, validCAPEM := generateTestCA(t, "valid-ca", 1, now.Add(-time.Hour), now.Add(time.Hour))
	_, _, expiredCAPEM := generateTestCA(t, "expired-ca", 2, now.Add(-2*time.Hour), now.Add(-time.Minute))

	validCert, validKey := writeTestLeafCert(t, dir, "valid-node", validCA, validCAKey, nil, now.Add(-time.Hour), now.Add(time.Hour))
	expiredCert, expiredKey := writeTestLeafCert(t, dir, "expired-node", validCA, validCAKey, nil, now.Add(-2*time.Hour), now.Add(-time.Minute))

	mixedBundle := filepath.Join(dir, "mixed-ca.crt")
	require.NoError(t, os.WriteFile(mixedBundle, slices.Concat(expiredCAPEM, validCAPEM), 0o600))

	expiredBundle := filepath.Join(dir, "expired-ca.crt")
	require.NoError(t, os.WriteFile(expiredBundle, expiredCAPEM, 0o600))

	corruptBundle := filepath.Join(dir, "corrupt-ca.crt")
	require.NoError(t, os.WriteFile(corruptBundle, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte("not-der")}), 0o600))

	t.Run("mixed expired and unexpired CAs succeeds and warns once", func(t *testing.T) {
		core, logs := observer.New(zapcore.DebugLevel)
		tlsInfo := NewTLSInfo(params.ConnectParameters{
			TLSCertFile: validCert,
			TLSKeyFile:  validKey,
			TLSCAFile:   mixedBundle,
		}, zap.New(core))

		require.NoError(t, checkCertificates(t.Context(), tlsInfo, zap.New(core)))
		assert.Equal(t, 1, logs.FilterMessage("raft: a CA of --raft_tls_ca_file has expired").Len())
	})

	t.Run("expired node certificate fails", func(t *testing.T) {
		tlsInfo := NewTLSInfo(params.ConnectParameters{
			TLSCertFile: expiredCert,
			TLSKeyFile:  expiredKey,
			TLSCAFile:   mixedBundle,
		}, zap.NewNop())

		err := checkCertificates(t.Context(), tlsInfo, zap.NewNop())
		require.ErrorContains(t, err, "expired on")
	})

	t.Run("all CAs expired fails", func(t *testing.T) {
		tlsInfo := NewTLSInfo(params.ConnectParameters{
			TLSCertFile: validCert,
			TLSKeyFile:  validKey,
			TLSCAFile:   expiredBundle,
		}, zap.NewNop())

		err := checkCertificates(t.Context(), tlsInfo, zap.NewNop())
		require.ErrorContains(t, err, "has no CA that has not expired")
	})

	t.Run("missing node certificate fails", func(t *testing.T) {
		tlsInfo := NewTLSInfo(params.ConnectParameters{
			TLSCertFile: filepath.Join(dir, "missing.crt"),
			TLSKeyFile:  validKey,
			TLSCAFile:   mixedBundle,
		}, zap.NewNop())

		require.Error(t, checkCertificates(t.Context(), tlsInfo, zap.NewNop()))
	})

	t.Run("missing or corrupt CA file fails", func(t *testing.T) {
		missingCA := NewTLSInfo(params.ConnectParameters{
			TLSCertFile: validCert,
			TLSKeyFile:  validKey,
			TLSCAFile:   filepath.Join(dir, "missing-ca.crt"),
		}, zap.NewNop())
		require.Error(t, checkCertificates(t.Context(), missingCA, zap.NewNop()))

		corruptCA := NewTLSInfo(params.ConnectParameters{
			TLSCertFile: validCert,
			TLSKeyFile:  validKey,
			TLSCAFile:   corruptBundle,
		}, zap.NewNop())
		require.Error(t, checkCertificates(t.Context(), corruptCA, zap.NewNop()))
	})
}

func TestMTLSListenerHandshakeAndPeerEndpoint(t *testing.T) {
	prevProvider := otel.GetMeterProvider()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(provider)
	t.Cleanup(func() {
		otel.SetMeterProvider(prevProvider)
		_ = provider.Shutdown(context.Background())
	})

	now := time.Now()
	dir := t.TempDir()

	trustedCA, trustedCAKey, trustedCAPEM := generateTestCA(t, "trusted-ca", 10, now.Add(-time.Hour), now.Add(time.Hour))
	caFile := filepath.Join(dir, "trusted-ca.crt")
	require.NoError(t, os.WriteFile(caFile, trustedCAPEM, 0o600))

	serverCert, serverKey := writeTestLeafCert(t, dir, "server", trustedCA, trustedCAKey, []net.IP{net.IPv4(127, 0, 0, 1)}, now.Add(-time.Hour), now.Add(time.Hour))
	// Client certificate intentionally omits IP SANs to verify SkipClientSANVerify on the listener.
	clientCert, clientKey := writeTestLeafCert(t, dir, "client-no-san", trustedCA, trustedCAKey, nil, now.Add(-time.Hour), now.Add(time.Hour))

	untrustedCA, untrustedCAKey, _ := generateTestCA(t, "untrusted-ca", 20, now.Add(-time.Hour), now.Add(time.Hour))
	untrustedCertPath, untrustedKeyPath := writeTestLeafCert(t, dir, "untrusted-peer", untrustedCA, untrustedCAKey, nil, now.Add(-time.Hour), now.Add(time.Hour))
	untrustedTLSCert, err := tls.LoadX509KeyPair(untrustedCertPath, untrustedKeyPath)
	require.NoError(t, err)

	core, logs := observer.New(zapcore.WarnLevel)
	logger := zap.New(core)

	serverTLS := NewTLSInfo(params.ConnectParameters{
		TLSCertFile: serverCert,
		TLSKeyFile:  serverKey,
		TLSCAFile:   caFile,
	}, logger)

	l, addr, err := startServerListener(t.Context(), logger, "127.0.0.1:0", serverTLS, func(err error) {
		t.Errorf("unexpected listener error: %v", err)
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = l.server.Shutdown(shutdownCtx)
	})

	peerURL := "https://" + addr
	wantInfo := node.Info{ID: 7, URL: peerURL, State: node.Waiting}
	serverEndpoint, err := newPeerEndpoint(logger, types.ID(42), serverTLS, func() (node.Info, error) {
		return wantInfo, nil
	})
	require.NoError(t, err)
	t.Cleanup(serverEndpoint.close)
	require.NoError(t, l.register(types.ID(42), serverEndpoint.handler(http.NotFoundHandler())))

	// 1. Trusted mTLS client succeeds even without IP SANs in its client certificate.
	clientTLS := NewTLSInfo(params.ConnectParameters{
		TLSCertFile: clientCert,
		TLSKeyFile:  clientKey,
		TLSCAFile:   caFile,
	}, zap.NewNop())
	clientEndpoint, err := newPeerEndpoint(zap.NewNop(), types.ID(42), clientTLS, func() (node.Info, error) {
		return node.Info{}, nil
	})
	require.NoError(t, err)
	t.Cleanup(clientEndpoint.close)

	gotInfo, err := clientEndpoint.fetch(t.Context(), peerURL)
	require.NoError(t, err)
	assert.Equal(t, wantInfo, gotInfo)

	// 2. Immediate TCP close (e.g. load-balancer probe) hits io.EOF and is ignored.
	eofConn, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", addr)
	require.NoError(t, err)
	require.NoError(t, eofConn.Close())

	// 3. Plaintext HTTP request triggers "not_tls".
	plainConn, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", addr)
	require.NoError(t, err)
	_, _ = plainConn.Write([]byte("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"))
	_ = plainConn.Close()

	// 4. Two rapid connections presenting an untrusted client cert both increment "unknown_authority",
	// while log rate-limiting suppresses the duplicate warning within the 5s window.
	for range 2 {
		dialer := &tls.Dialer{
			Config: &tls.Config{
				GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
					return &untrustedTLSCert, nil
				},
				InsecureSkipVerify: true,
			},
		}
		conn, dialErr := dialer.DialContext(t.Context(), "tcp", addr)
		if dialErr == nil {
			_, _ = conn.Write([]byte("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"))
			var buf [1]byte
			_, _ = conn.Read(buf[:])
			_ = conn.Close()
		}
	}

	failureCounts := func() map[string]int64 {
		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(t.Context(), &rm))
		counts := map[string]int64{}
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				if m.Name != "raft_peer_tls_handshake_failures" {
					continue
				}
				sum, ok := m.Data.(metricdata.Sum[int64])
				if !ok {
					continue
				}
				for _, dp := range sum.DataPoints {
					if reason, ok := dp.Attributes.Value(attribute.Key("reason")); ok {
						counts[reason.AsString()] = dp.Value
					}
				}
			}
		}
		return counts
	}

	require.Eventually(t, func() bool {
		counts := failureCounts()
		return counts["not_tls"] == 1 && counts["unknown_authority"] == 2
	}, 5*time.Second, 10*time.Millisecond)

	warnEntries := logs.FilterMessage("raft: tls handshake failure").All()
	var unknownAuthLogs []observer.LoggedEntry
	for _, entry := range warnEntries {
		if entry.ContextMap()["reason"] == "unknown_authority" {
			unknownAuthLogs = append(unknownAuthLogs, entry)
		}
	}
	require.Len(t, unknownAuthLogs, 1, "second unknown_authority failure within 5s should be rate-limited")
	assert.Equal(t, "CN=untrusted-peer", unknownAuthLogs[0].ContextMap()["peer_subject"])
}
