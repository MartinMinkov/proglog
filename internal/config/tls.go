package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

type TLSConfig struct {
	CertFile      string
	KeyFile       string
	CAFile        string
	ServerAddress string
	Server        bool
}

func SetupTLSConfig(config TLSConfig) (*tls.Config, error) {
	var err error
	tlsConfig := &tls.Config{}
	if config.CertFile != "" && config.KeyFile != "" {
		tlsConfig.Certificates = make([]tls.Certificate, 1)
		// Load the certificate and key from the provided files
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, err
		}
	}
	if config.CAFile != "" {
		// Load the CA certificate from the provided file.
		b, err := os.ReadFile(config.CAFile)
		if err != nil {
			return nil, err
		}
		// Create a new certicfcate pool to store the CA certificate
		ca := x509.NewCertPool()
		// Append the CA certificate to the pool
		ok := ca.AppendCertsFromPEM(b)
		if !ok {
			return nil, fmt.Errorf("failed to parse root CA certificate: %w", err)
		}
		if config.Server {
			// If the server is true, we set the ClientCAs field to the CA pool and set the ClientAuth field to RequireAndVerifyClientCert
			tlsConfig.ClientCAs = ca
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			// If the server is false, we set the RootCAs field to the CA pool
			tlsConfig.RootCAs = ca
		}
		// Set the ServerName field to the server address
		tlsConfig.ServerName = config.ServerAddress
	}
	return tlsConfig, nil
}
