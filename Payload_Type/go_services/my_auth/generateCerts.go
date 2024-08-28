package my_auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return !info.IsDir()
}

func generateCerts(name string) error {
	certPath := fmt.Sprintf("%s.crt", name)
	keyPath := fmt.Sprintf("%s.key", name)
	if FileExists(certPath) && FileExists(keyPath) {
		return nil
	}
	fmt.Printf("[*] Failed to find SSL certs for Nginx container, generating now...\n")
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	notBefore := time.Now()
	oneYear := 365 * 24 * time.Hour
	notAfter := notBefore.Add(oneYear)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Printf("[-] failed to generate serial number: %s\n", err)
		return err
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Mythic"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		fmt.Printf("[-] Failed to create certificate: %s\n", err)
		return err
	}
	certOut, err := os.Create(certPath)
	if err != nil {
		fmt.Printf("[-] failed to open "+certPath+" for writing: %s\n", err)
		return err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	certOut.Close()
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	marshalKey := x509.MarshalPKCS1PrivateKey(priv)

	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: marshalKey})
	keyOut.Close()
	fmt.Printf("[+] Successfully generated new SSL certs\n")
	return nil
}
