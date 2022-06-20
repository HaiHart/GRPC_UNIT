package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// SSLCerts represents the required certificate files for interacting with the BDN.
type SSLCerts struct {
	privateCertFile string
	privateKeyFile  string

	privateCert    *x509.Certificate
	privateKey     ecdsa.PrivateKey
	privateKeyPair *tls.Certificate
}

// GetCertDir getting cert, key and registration files
func GetCertDir(privateBaseURL, certName string) (privateCertFile string, privateKeyFile string) {
	privateDir := path.Join(privateBaseURL, certName, "private")
	_ = os.MkdirAll(privateDir, 0755)
	privateCertFile = path.Join(privateDir, fmt.Sprintf("%v_cert.pem", certName))
	privateKeyFile = path.Join(privateDir, fmt.Sprintf("%v_key.pem", certName))
	return
}

// NewSSLCertsFromFiles receiving cert files info returns and initializes new storage of SSL certificates.
// Private keys and certs must match each other. If they do not, a new private key will be generated
// and written, pending loading of a new certificate.
func NewSSLCertsFromFiles(privateCertFile, privateKeyFile string) SSLCerts {
	var privateKey *ecdsa.PrivateKey
	var privateCert *x509.Certificate
	var privateKeyPair *tls.Certificate

	privateCertBlock, err := ioutil.ReadFile(privateCertFile)
	if err != nil {
		privateCertBlock = nil
	}
	privateKeyBlock, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		privateKeyBlock = nil
	}

	if privateKeyBlock == nil && privateCertBlock == nil {
		privateKey, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
		if err != nil {
			panic(fmt.Errorf("could not generate private key: %v", err))
		}
		privateKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
		if err != nil {
			panic(fmt.Errorf("could not marshal generated private key: %v", err))
		}
		privateKeyBlock = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privateKeyBytes})
		err = ioutil.WriteFile(privateKeyFile, privateCertBlock, 0644)
		if err != nil {
			panic(fmt.Errorf("could not write new private key to file (%v): %v", privateKeyFile, err))
		}
	} else if privateKeyBlock != nil && privateCertBlock != nil {
		privateKey, err = parsePEMPrivateKey(privateKeyBlock)
		if err != nil {
			panic(fmt.Errorf("could not parse private key from file (%v): %v", privateKeyFile, err))
		}
		privateCert, err = parsePEMCert(privateCertBlock)
		if err != nil {
			panic(fmt.Errorf("could not parse private cert from file (%v): %v", privateCert, err))
		}
		_privateKeyPair, err := tls.X509KeyPair(privateCertBlock, privateKeyBlock)
		if err != nil {
			panic(fmt.Errorf("could not load private key pair: %v", err))
		}
		privateKeyPair = &_privateKeyPair
	} else if privateKeyBlock != nil {
		privateKey, err = parsePEMPrivateKey(privateKeyBlock)
		if err != nil {
			panic(fmt.Errorf("could not parse private key from file (%v): %v", privateKeyFile, err))
		}
	} else {
		panic(fmt.Errorf("found a certificate with no matching private key –– delete the certificate at %v if it's not needed", privateCertFile))
	}

	return SSLCerts{
		privateCertFile: privateCertFile,
		privateKeyFile:  privateKeyFile,
		privateCert:     privateCert,
		privateKey:      *privateKey,
		privateKeyPair:  privateKeyPair,
	}
}

// parsePEMPrivateKey should parse pem private key
func parsePEMPrivateKey(block []byte) (*ecdsa.PrivateKey, error) {
	decodedKeyBlock, _ := pem.Decode(block)
	keyBytes := decodedKeyBlock.Bytes
	return x509.ParseECPrivateKey(keyBytes)
}

func parsePEMCert(block []byte) (*x509.Certificate, error) {
	decodedCertBlock, _ := pem.Decode(block)
	certBytes := decodedCertBlock.Bytes
	return x509.ParseCertificate(certBytes)
}

// LoadPrivateConfig generates TLS config from the private certificates.
func (s SSLCerts) LoadPrivateConfig() (*tls.Config, error) {
	if s.privateKeyPair == nil {
		return nil, errors.New("private key pair has not been loaded")
	}
	config := &tls.Config{
		Certificates:       []tls.Certificate{*s.privateKeyPair},
		InsecureSkipVerify: true,
	}
	return config, nil
}
