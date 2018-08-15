package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/reconquest/karma-go"
)

const (
	secureLayerCertificatePath = "tls/cert.pem"
	secureLayerKeyPath         = "tls/key.pem"
)

type SecureLayer struct {
	Certificate     *x509.Certificate
	CertificatePEM  []byte
	CertificatePath string
	Key             *rsa.PrivateKey
	KeyPath         string
	Fingerprint     Fingerprint
	X509KeyPair     tls.Certificate
}

type Fingerprint []byte

func (fingerprint Fingerprint) String() string {
	buffer := bytes.NewBuffer(nil)
	for i, _ := range fingerprint {
		if i != 0 {
			buffer.WriteRune(':')
		}

		chunk := fingerprint[i : i+1]
		buffer.WriteString(
			strings.ToUpper(hex.EncodeToString(chunk)),
		)
	}

	return buffer.String()
}

func getSecureLayer(dataDir string, withKey bool) (*SecureLayer, error) {
	layer := SecureLayer{}
	layer.CertificatePath = filepath.Join(dataDir, secureLayerCertificatePath)
	layer.KeyPath = filepath.Join(dataDir, secureLayerKeyPath)

	found := true
	for _, path := range []string{
		layer.CertificatePath,
		layer.KeyPath,
	} {
		_, err := os.Stat(path)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}

		if os.IsNotExist(err) {
			found = false
			break
		}
	}

	if !found {
		logger.Infof("generating TLS certificate")

		// 10 years
		invalidAfter := time.Now().AddDate(10, 0, 0)

		err := generateCertificate(dataDir, 4096, invalidAfter)
		if err != nil {
			return nil, karma.Format(
				err,
				"unable to generate TLS key-cert pair",
			)
		}
	}

	certPEM, err := ioutil.ReadFile(layer.CertificatePath)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to read certificate file: %s", layer.CertificatePath,
		)
	}

	layer.CertificatePEM = certPEM

	keyPEM, err := ioutil.ReadFile(layer.KeyPath)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to read key file: %s", layer.KeyPath,
		)
	}

	certBlock, _ := pem.Decode([]byte(certPEM))
	if certBlock == nil {
		return nil, karma.Format(
			err,
			"unable to decode certificate PEM data",
		)
	}

	keyBlock, _ := pem.Decode([]byte(keyPEM))
	if keyBlock == nil {
		return nil, karma.Format(
			err,
			"unable to decode key PEM data",
		)
	}

	layer.X509KeyPair, err = tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to parse X509 key pair",
		)
	}

	layer.Certificate, err = x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to parse certificate PEM block",
		)
	}

	layer.Key, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		return nil, karma.Format(
			err,
			"unable to parse certificate PEM block",
		)
	}

	layer.Fingerprint = getFingerprint(layer.Certificate)

	return &layer, nil
}

func getFingerprint(cert *x509.Certificate) Fingerprint {
	hasher := sha1.New()
	hasher.Write(cert.Raw)

	return Fingerprint(hasher.Sum(nil))
}

func generateCertificate(
	certDir string, blockSize int, invalidAfter time.Time,
) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, blockSize)
	if err != nil {
		return karma.Format(
			err,
			"unable to generate RSA key",
		)
	}

	invalidBefore := time.Now()

	serialNumberBlockSize := big.NewInt(0).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberBlockSize)
	if err != nil {
		return karma.Format(err, "failed to generate serial number")
	}

	cert := x509.Certificate{
		IsCA: true,

		SerialNumber: serialNumber,

		NotBefore: invalidBefore,
		NotAfter:  invalidAfter,

		BasicConstraintsValid: true,
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},

		Subject: pkix.Name{CommonName: "monk"},
	}

	certData, err := x509.CreateCertificate(
		rand.Reader, &cert, &cert, &privateKey.PublicKey, privateKey,
	)
	if err != nil {
		return karma.Format(
			err, "can't create certificate",
		)
	}

	certOutFd, err := os.Create(filepath.Join(certDir, secureLayerCertificatePath))
	if err != nil {
		return karma.Format(
			err, "can't create certificate file",
		)
	}

	err = pem.Encode(
		certOutFd,
		&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: certData,
		},
	)
	if err != nil {
		return karma.Format(
			err, "can't write PEM data to certificate file",
		)
	}

	err = certOutFd.Close()
	if err != nil {
		return karma.Format(
			err, "can't close certificate file",
		)
	}

	keyOutFd, err := os.OpenFile(
		filepath.Join(certDir, secureLayerKeyPath),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return karma.Format(
			err, "can't open key file",
		)
	}

	err = pem.Encode(
		keyOutFd,
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
		},
	)
	if err != nil {
		return karma.Format(
			err, "can't write PEM data to key file",
		)
	}

	err = keyOutFd.Close()
	if err != nil {
		return karma.Format(
			err, "can't close key file",
		)
	}

	return nil
}
