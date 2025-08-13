package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"
)

// symmetric key for AES (32 bytes for AES-256)
// In production, generate securely and share via secure channel
var encryptionKey = []byte("0123456789abcdef0123456789abcdef")

// Load your CA cert and key from files (PEM format)
// func loadCA(keyPathth, certPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
// 	caCertPEM, err := os.ReadFile(certPath + "/ca.crt")
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	caKeyPEM, err := os.ReadFile(keyPathth + "/ca.key")
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	fmt.Println("caCertPEM: ", caCertPEM)
// 	block, _ := pem.Decode(caCertPEM)
// 	if block == nil {
// 		return nil, nil, fmt.Errorf("failed to parse CA cert PEM")
// 	}
// 	caCert, err := x509.ParseCertificate(block.Bytes)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	block, _ = pem.Decode(caKeyPEM)
// 	if block == nil {
// 		return nil, nil, fmt.Errorf("failed to parse CA key PEM")
// 	}
// 	caKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	return caCert, caKey, nil
// }

func loadCA(keyPathth, certPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	caCertPEM, err := os.ReadFile(certPath + "/ca.crt")
	if err != nil {
		return nil, nil, err
	}
	caKeyPEM, err := os.ReadFile(keyPathth + "/ca.key")
	if err != nil {
		return nil, nil, err
	}

	block, _ := pem.Decode(caCertPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to parse CA cert PEM")
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, err
	}

	block, _ = pem.Decode(caKeyPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to parse CA key PEM")
	}

	// Try to parse PKCS#8 private key
	keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		// fallback: try PKCS#1 parsing
		caKey, err2 := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, nil, fmt.Errorf("failed to parse private key: %v / %v", err, err2)
		}
		return caCert, caKey, nil
	}

	caKey, ok := keyInterface.(*rsa.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("private key is not RSA")
	}

	return caCert, caKey, nil
}

// Generate a client cert + private key signed by CA
func generateClientCert(caCert *x509.Certificate, caKey *rsa.PrivateKey, commonName string) ([]byte, []byte, error) {
	// Generate client private key
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Client cert template
	clientCertTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour), // 1 year validity

		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}

	// fmt.Println("clientKey.PublicKey: ", caKey)
	// Create cert signed by CA
	certDER, err := x509.CreateCertificate(rand.Reader, clientCertTmpl, caCert, &clientKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	// Encode private key to PEM
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientKey),
	})

	// Encode cert to PEM
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDER,
	})

	return certPEM, keyPEM, nil
}

func ProceedCerts(vpsID string) (map[string]string, error) {
	path := "./certs"
	caCert, caKey, err := loadCA(path, path)
	if err != nil {
		fmt.Println("Error: ", err)
		// panic(err)
	}

	clientCN := "Agent_" + vpsID

	// fmt.Println("caCert:", caCert)
	// fmt.Println("caKey:", caKey)

	clientCertPEM, clientKeyPEM, err := generateClientCert(caCert, caKey, clientCN)
	if err != nil {
		panic(err)
	}

	// Save to disk for demo (in real use, send to client)
	// os.WriteFile("client.crt", clientCertPEM, 0644)
	// os.WriteFile("client.key", clientKeyPEM, 0600)

	fmt.Println("Generated client certificate and key for:", clientCN)

	keys := map[string]string{
		"cert": string(clientCertPEM),
		"key":  string(clientKeyPEM),
	}

	return keys, nil

}

// Encrypt data with AES-GCM symmetric key
func encryptAESGCM(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}
