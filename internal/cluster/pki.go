package cluster

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ─── PKI (Public Key Infrastructure) for mTLS ─────────────────────────

// PKIConfig holds paths and settings for certificate management.
type PKIConfig struct {
	CADir       string // directory to store CA cert + key
	NodeDir     string // directory to store node cert + key
	Org         string // organization name for certs
	Validity    time.Duration
	GracePeriod time.Duration // how long before expiry to auto-renew
}

// PKIManager handles certificate lifecycle for the cluster mTLS mesh.
type PKIManager struct {
	cfg           PKIConfig
	caCert        *x509.Certificate
	caKey         any
	nodeCert      *x509.Certificate
	nodeKey       any
	nodeCertPEM   []byte
	nodeKeyPEM    []byte
	caCertPEM     []byte
	renewTicker   *time.Ticker
	stopCh        chan struct{}
}

// NewPKIManager creates a PKI manager and loads or generates certs.
func NewPKIManager(cfg PKIConfig) (*PKIManager, error) {
	if cfg.CADir == "" {
		cfg.CADir = "/etc/vortex/pki/ca"
	}
	if cfg.NodeDir == "" {
		cfg.NodeDir = "/etc/vortex/pki/node"
	}
	if cfg.Org == "" {
		cfg.Org = "VortexUiPro"
	}
	if cfg.Validity <= 0 {
		cfg.Validity = 365 * 24 * time.Hour // 1 year
	}
	if cfg.GracePeriod <= 0 {
		cfg.GracePeriod = 30 * 24 * time.Hour // 30 days
	}

	os.MkdirAll(cfg.CADir, 0700)
	os.MkdirAll(cfg.NodeDir, 0700)

	p := &PKIManager{
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}

	if err := p.loadOrGenerate(); err != nil {
		return nil, fmt.Errorf("pki setup: %w", err)
	}

	return p, nil
}

func (p *PKIManager) loadOrGenerate() error {
	// Try to load existing CA
	caCertPath := filepath.Join(p.cfg.CADir, "ca.pem")
	caKeyPath := filepath.Join(p.cfg.CADir, "ca-key.pem")

	if p.loadCA(caCertPath, caKeyPath) {
		log.Println("PKI: Loaded existing CA certificate")
	} else {
		if err := p.generateCA(caCertPath, caKeyPath); err != nil {
			return fmt.Errorf("generate CA: %w", err)
		}
		log.Println("PKI: Generated new CA certificate")
	}

	// Try to load existing node cert
	nodeCertPath := filepath.Join(p.cfg.NodeDir, "node.pem")
	nodeKeyPath := filepath.Join(p.cfg.NodeDir, "node-key.pem")

	if p.loadNodeCert(nodeCertPath, nodeKeyPath) {
		log.Println("PKI: Loaded existing node certificate")
		// Check if renewal is needed
		if p.nodeCert != nil {
			remaining := time.Until(p.nodeCert.NotAfter)
			if remaining < p.cfg.GracePeriod {
				log.Printf("PKI: Node cert expires in %v, renewing...", remaining)
				if err := p.generateNodeCert(nodeCertPath, nodeKeyPath); err != nil {
					return fmt.Errorf("renew node cert: %w", err)
				}
			}
		}
	} else {
		if err := p.generateNodeCert(nodeCertPath, nodeKeyPath); err != nil {
			return fmt.Errorf("generate node cert: %w", err)
		}
		log.Println("PKI: Generated new node certificate")
	}

	return nil
}

// StartAutoRenew begins periodic certificate renewal checks.
func (p *PKIManager) StartAutoRenew() {
	p.renewTicker = time.NewTicker(24 * time.Hour)
	go func() {
		for {
			select {
			case <-p.renewTicker.C:
				p.checkAndRenew()
			case <-p.stopCh:
				return
			}
		}
	}()
	log.Println("PKI: Auto-renewal started (check every 24h)")
}

// Stop stops the auto-renewal loop.
func (p *PKIManager) Stop() {
	if p.renewTicker != nil {
		p.renewTicker.Stop()
	}
	close(p.stopCh)
}

func (p *PKIManager) checkAndRenew() {
	if p.nodeCert == nil {
		return
	}
	remaining := time.Until(p.nodeCert.NotAfter)
	if remaining < p.cfg.GracePeriod {
		log.Printf("PKI: Node cert expires in %v, renewing...", remaining)
		nodeCertPath := filepath.Join(p.cfg.NodeDir, "node.pem")
		nodeKeyPath := filepath.Join(p.cfg.NodeDir, "node-key.pem")
		if err := p.generateNodeCert(nodeCertPath, nodeKeyPath); err != nil {
			log.Printf("PKI: Renewal failed: %v", err)
		} else {
			log.Println("PKI: Node certificate renewed successfully")
		}
	}
}

// ─── CA Generation ───────────────────────────────────────────────────

func (p *PKIManager) generateCA(certPath, keyPath string) error {
	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("generate CA key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("generate CA serial: %w", err)
	}

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{p.cfg.Org},
			CommonName:   p.cfg.Org + " Cluster CA",
		},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(p.cfg.Validity * 10), // CA lasts 10x longer
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return fmt.Errorf("create CA cert: %w", err)
	}

	p.caCert, err = x509.ParseCertificate(certDER)
	if err != nil {
		return fmt.Errorf("parse CA cert: %w", err)
	}
	p.caKey = key

	// Write PEM files
	if err := writePEM(certPath, "CERTIFICATE", certDER, 0644); err != nil {
		return err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("marshal CA key: %w", err)
	}
	if err := writePEM(keyPath, "PRIVATE KEY", keyDER, 0600); err != nil {
		return err
	}

	p.caCertPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	return nil
}

func (p *PKIManager) loadCA(certPath, keyPath string) bool {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return false
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return false
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return false
	}
	p.caCert, err = x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return false
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return false
	}
	p.caKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		p.caKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			return false
		}
	}

	p.caCertPEM = certPEM
	return true
}

// ─── Node Certificate Generation ─────────────────────────────────────

func (p *PKIManager) generateNodeCert(certPath, keyPath string) error {
	if p.caCert == nil || p.caKey == nil {
		return fmt.Errorf("CA not initialized")
	}

	// Generate ECDSA key for node (faster, smaller)
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate node key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("generate serial: %w", err)
	}

	certHash := sha256.Sum256([]byte(p.cfg.Org + "-node-" + time.Now().String()))

	template := &x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{p.cfg.Org},
			CommonName:   p.cfg.Org + " Cluster Node",
		},
		NotBefore: time.Now().Add(-1 * time.Hour),
		NotAfter:  time.Now().Add(p.cfg.Validity),

		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},

		BasicConstraintsValid: true,
		IsCA:                  false,

		// Subject Key Identifier for CA verification
		SubjectKeyId: certHash[:],
	}

	// Add SANs for all interfaces and common names
	template.DNSNames = []string{"localhost", "node.local", "cluster.local"}
	template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}

	// Sign with CA
	certDER, err := x509.CreateCertificate(rand.Reader, template, p.caCert, &key.PublicKey, p.caKey)
	if err != nil {
		return fmt.Errorf("create node cert: %w", err)
	}

	p.nodeCert, err = x509.ParseCertificate(certDER)
	if err != nil {
		return fmt.Errorf("parse node cert: %w", err)
	}
	p.nodeKey = key

	// Write PEM files
	if err := writePEM(certPath, "CERTIFICATE", certDER, 0644); err != nil {
		return err
	}
	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		return fmt.Errorf("marshal node key: %w", err)
	}
	if err := writePEM(keyPath, "PRIVATE KEY", keyDER, 0600); err != nil {
		return err
	}

	p.nodeCertPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	p.nodeKeyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	return nil
}

func (p *PKIManager) loadNodeCert(certPath, keyPath string) bool {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return false
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return false
	}

	certBlock, _ := pem.Decode(certPEM)
	if certBlock == nil {
		return false
	}
	p.nodeCert, err = x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return false
	}

	keyBlock, _ := pem.Decode(keyPEM)
	if keyBlock == nil {
		return false
	}
	p.nodeKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		return false
	}

	p.nodeCertPEM = certPEM
	p.nodeKeyPEM = keyPEM
	return true
}

// ─── Accessors ───────────────────────────────────────────────────────

// CACertPool returns the CA certificate pool for TLS verification.
func (p *PKIManager) CACertPool() *x509.CertPool {
	pool := x509.NewCertPool()
	if p.caCertPEM != nil {
		pool.AppendCertsFromPEM(p.caCertPEM)
	}
	return pool
}

// NodeCertPEM returns the PEM-encoded node certificate.
func (p *PKIManager) NodeCertPEM() []byte { return p.nodeCertPEM }

// NodeKeyPEM returns the PEM-encoded node private key.
func (p *PKIManager) NodeKeyPEM() []byte { return p.nodeKeyPEM }

// CACertPEM returns the PEM-encoded CA certificate.
func (p *PKIManager) CACertPEM() []byte { return p.caCertPEM }

// CertExpiry returns the node certificate expiry time.
func (p *PKIManager) CertExpiry() time.Time {
	if p.nodeCert == nil {
		return time.Time{}
	}
	return p.nodeCert.NotAfter
}

// Stats returns PKI status information.
func (p *PKIManager) Stats() map[string]any {
	certInfo := map[string]any{}
	if p.nodeCert != nil {
		certInfo = map[string]any{
			"not_before": p.nodeCert.NotBefore.UnixMilli(),
			"not_after":  p.nodeCert.NotAfter.UnixMilli(),
			"expires_in": time.Until(p.nodeCert.NotAfter).String(),
		}
	}
	return map[string]any{
		"ca_exists":  p.caCert != nil,
		"cert_exists": p.nodeCert != nil,
		"cert":       certInfo,
		"ca_dir":     p.cfg.CADir,
		"node_dir":   p.cfg.NodeDir,
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────

func writePEM(path, blockType string, derBytes []byte, mode os.FileMode) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := pem.Encode(f, &pem.Block{Type: blockType, Bytes: derBytes}); err != nil {
		return err
	}
	return f.Chmod(mode)
}

// SanitizeAddrForCert strips port and brackets from an address for use in cert SANs.
func SanitizeAddrForCert(addr string) string {
	// Strip port
	if idx := strings.LastIndex(addr, ":"); idx >= 0 {
		addr = addr[:idx]
	}
	// Strip brackets (IPv6)
	addr = strings.Trim(addr, "[]")
	return addr
}
