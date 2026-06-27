package trust

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbeSystemTrust_selfSigned(t *testing.T) {
	srv := startHTTPServer(t, "127.0.0.1", []string{"127.0.0.1"}, []net.IP{net.ParseIP("127.0.0.1")})
	defer srv.Close()

	host, port, err := net.SplitHostPort(srv.Listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	trusted, err := ProbeSystemTrust(context.Background(), host, port)
	if trusted {
		t.Fatal("expected self-signed cert to fail system trust")
	}
	if !IsUnknownAuthority(err) {
		t.Fatalf("expected unknown authority, got %v", err)
	}
}

func TestIsUnknownAuthorityAndHostnameMismatch(t *testing.T) {
	if IsUnknownAuthority(nil) || IsHostnameMismatch(nil) {
		t.Fatal("nil should be false")
	}
	if !IsUnknownAuthority(errors.New("x509: certificate signed by unknown authority")) {
		t.Fatal("string match unknown authority")
	}
	if !IsHostnameMismatch(errors.New(`x509: certificate is valid for example.com, not localhost`)) {
		t.Fatal("string match hostname")
	}
}

func startHTTPServer(t *testing.T, cn string, dnsNames []string, ips []net.IP) *httptest.Server {
	t.Helper()
	tlsCert := selfSignedTLSCert(t, cn, dnsNames, ips)
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	srv.TLS = &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		MinVersion:   tls.VersionTLS12,
	}
	srv.StartTLS()
	return srv
}

func selfSignedTLSCert(t *testing.T, cn string, dnsNames []string, ips []net.IP) tls.Certificate {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: cn},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		DNSNames:     dnsNames,
		IPAddresses:  ips,
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		t.Fatal(err)
	}
	return tlsCert
}
