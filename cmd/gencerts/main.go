package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ─── mTLS Certificate Generator for VortexUiPro ──────────────────────
// Generates a self-signed CA + server/client certificates for cluster
// mTLS mesh communication.
//
// Usage:
//   go run ./cmd/gencerts -out ./deploy/certs -san localhost,127.0.0.1,node-1
// ──────────────────────────────────────────────────────────────────────

var (
	outputDir = flag.String("out", "./deploy/certs", "Output directory for certificates")
	sans      = flag.String("san", "localhost,127.0.0.1", "Comma-separated SANs (DNS names and IPs)")
	validity  = flag.Duration("validity", 365*24*time.Hour, "Certificate validity duration")
)

func main() {
	flag.Parse()

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	dnsNames, ipAddresses := parseSANs(*sans)

	// Generate CA key and certificate
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate CA key: %v", err)
	}

	caTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "VortexUiPro CA"},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(*validity),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		log.Fatalf("Failed to create CA certificate: %v", err)
	}

	writePEM(filepath.Join(*outputDir, "ca.crt"), "CERTIFICATE", caDER)
	writeKey(filepath.Join(*outputDir, "ca.key"), caKey)
	log.Println("✅ CA certificate generated")

	// Generate server certificate
	serverKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate server key: %v", err)
	}

	serverTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "VortexUiPro Server"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(*validity),
		DNSNames:     dnsNames,
		IPAddresses:  ipAddresses,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, serverTemplate, caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		log.Fatalf("Failed to create server certificate: %v", err)
	}

	writePEM(filepath.Join(*outputDir, "panel.crt"), "CERTIFICATE", serverDER)
	writeKey(filepath.Join(*outputDir, "panel.key"), serverKey)
	// Copy as node cert/key too for convenience
	copyFile(filepath.Join(*outputDir, "panel.crt"), filepath.Join(*outputDir, "node.crt"))
	copyFile(filepath.Join(*outputDir, "panel.key"), filepath.Join(*outputDir, "node.key"))
	log.Println("✅ Server certificates generated")

	fmt.Println()
	fmt.Println("📁 Certificates generated in:", *outputDir)
	fmt.Println("   ├── ca.crt     (CA certificate)")
	fmt.Println("   ├── ca.key     (CA private key)")
	fmt.Println("   ├── panel.crt  (Server certificate)")
	fmt.Println("   ├── panel.key  (Server private key)")
	fmt.Println("   ├── node.crt   (Node certificate, copy of panel.crt)")
	fmt.Println("   └── node.key   (Node private key, copy of panel.key)")
	fmt.Println()
	fmt.Printf("🔒 Valid until: %s\n", time.Now().Add(*validity).Format(time.RFC3339))
	fmt.Println("📌 To use with VortexUiPro cluster, set env vars:")
	fmt.Println("   VORTEX_CLUSTER_TLS_CERT=" + filepath.Join(*outputDir, "panel.crt"))
	fmt.Println("   VORTEX_CLUSTER_TLS_KEY=" + filepath.Join(*outputDir, "panel.key"))
	fmt.Println("   VORTEX_CLUSTER_TLS_CA=" + filepath.Join(*outputDir, "ca.crt"))
}

// ─── Helpers ─────────────────────────────────────────────────────────

func parseSANs(s string) (dnsNames []string, ipAddresses []net.IP) {
	for _, entry := range strings.Split(s, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if ip := net.ParseIP(entry); ip != nil {
			ipAddresses = append(ipAddresses, ip)
		} else {
			dnsNames = append(dnsNames, entry)
		}
	}
	return
}

func writePEM(path, blockType string, derBytes []byte) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("Failed to write %s: %v", path, err)
	}
	defer f.Close()
	pem.Encode(f, &pem.Block{Type: blockType, Bytes: derBytes})
}

func writeKey(path string, key *ecdsa.PrivateKey) {
	b, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		log.Fatalf("Failed to marshal key: %v", err)
	}
	writePEM(path, "EC PRIVATE KEY", b)
}

func copyFile(src, dst string) {
	data, err := os.ReadFile(src)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", src, err)
	}
	if err := os.WriteFile(dst, data, 0644); err != nil {
		log.Fatalf("Failed to write %s: %v", dst, err)
	}
}
