package config

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type TLSConfig struct {
	// Enabled determines if MQTT connection should be over TLS
	Enabled bool

	// RootCAFile is the root CA certificate used to verify server certificate.
	// This is optional if SkipVerify is true.
	// If SkipVerify is false and RootCAFile is not set then the default openssl
	// CA bundle will be used.
	RootCAFile string

	// ClientCertFile is the client X509 certificate file
	ClientCertFile string

	// ClientKeyFile is the client RSA key file
	ClientKeyFile string

	// SkipVerify determines if certificate contents shouldn't
	// be matched against server.
	SkipVerify bool
}

func DefaultTLSConfig() TLSConfig {
	return TLSConfig{
		Enabled:        false,
		RootCAFile:     "",
		ClientCertFile: "",
		ClientKeyFile:  "",
		SkipVerify:     false,
	}
}

// Load loads the certificate/key pair defined inside the TLSConfig
// and returns the loaded std tls.Config
func (c *TLSConfig) Load() (*tls.Config, error) {
	if !c.Enabled {
		// Return a nil configuration if TLS is disabled
		return nil, nil
	}

	certPool := x509.NewCertPool()

	if c.RootCAFile != "" {
		pemCerts, err := ioutil.ReadFile(c.RootCAFile)
		if err != nil {
			return nil, eris.Wrap(err, "failed to read Root CA file")
		}
		certPool.AppendCertsFromPEM(pemCerts)
	}

	// Import client certificate/key pair
	cert, err := tls.LoadX509KeyPair(
		c.ClientCertFile,
		c.ClientKeyFile,
	)
	if err != nil {
		return nil, eris.Wrap(err, "failed to read client cert/key pair")
	}

	// Immediately parse certificate
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, eris.Wrap(err, "failed to parse client certificate")
	}

	if log.IsLevelEnabled(log.TraceLevel) {
		log.Trace("TLS cert: ", cert.Leaf)
	}

	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: c.SkipVerify,
		Certificates:       []tls.Certificate{cert},
	}

	return tlsConfig, nil
}
