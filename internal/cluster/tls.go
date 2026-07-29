package cluster

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
)

// ─── Real mTLS Configuration (uses crypto/tls) ───────────────────────

// TLSConfig returns a *tls.Config for the peer server using the PKI manager.
func TLSConfig(pki *PKIManager, isServer bool) (*tls.Config, error) {
	if pki == nil {
		return nil, fmt.Errorf("PKI manager required for TLS config")
	}

	cert, err := tls.X509KeyPair(pki.NodeCertPEM(), pki.NodeKeyPEM())
	if err != nil {
		return nil, fmt.Errorf("load key pair: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(pki.CACertPEM())

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ClientCAs:    caCertPool,
		MinVersion:   tls.VersionTLS12,
		ServerName:   "cluster.local",
	}

	if isServer {
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return cfg, nil
}

// TLSDialConfig returns a TLS config for client-side connections.
func TLSDialConfig(pki *PKIManager, serverName string) (*tls.Config, error) {
	if pki == nil {
		return nil, fmt.Errorf("PKI manager required for TLS config")
	}

	cert, err := tls.X509KeyPair(pki.NodeCertPEM(), pki.NodeKeyPEM())
	if err != nil {
		return nil, fmt.Errorf("load key pair: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(pki.CACertPEM())

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
		ServerName:   serverName,
		MinVersion:   tls.VersionTLS12,
	}, nil
}
