package run

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/mongodb-forks/drone-helm3/internal/env"
)

type repoCerts struct {
	*config
	cert           string
	certFilename   string
	caCert         string
	caCertFilename string
}

func newRepoCerts(cfg env.Config) *repoCerts {
	return &repoCerts{
		config: newConfig(cfg),
		cert:   cfg.RepoCertificate,
		caCert: cfg.RepoCACertificate,
	}
}

func (rc *repoCerts) write() error {
	if rc.cert != "" {
		file, err := os.CreateTemp("", "repo********.cert")
		if err != nil {
			return fmt.Errorf("failed to create certificate file: %w", err)
		}
		defer file.Close()

		rc.certFilename = file.Name()
		rawCert, err := base64.StdEncoding.DecodeString(rc.cert)
		if err != nil {
			return fmt.Errorf("failed to base64-decode certificate string: %w", err)
		}
		if rc.debug {
			fmt.Fprintf(rc.stderr, "writing repo certificate to %s\n", rc.certFilename)
		}
		if _, err := file.Write(rawCert); err != nil {
			return fmt.Errorf("failed to write certificate file: %w", err)
		}
	}

	if rc.caCert != "" {
		file, err := os.CreateTemp("", "repo********.ca.cert")
		if err != nil {
			return fmt.Errorf("failed to create CA certificate file: %w", err)
		}
		defer file.Close()

		rc.caCertFilename = file.Name()
		rawCert, err := base64.StdEncoding.DecodeString(rc.caCert)
		if err != nil {
			return fmt.Errorf("failed to base64-decode CA certificate string: %w", err)
		}
		if rc.debug {
			fmt.Fprintf(rc.stderr, "writing repo ca certificate to %s\n", rc.caCertFilename)
		}
		if _, err := file.Write(rawCert); err != nil {
			return fmt.Errorf("failed to write CA certificate file: %w", err)
		}
	}
	return nil
}

func (rc *repoCerts) flags() []string {
	flags := make([]string, 0)
	if rc.certFilename != "" {
		flags = append(flags, "--cert-file", rc.certFilename)
	}
	if rc.caCertFilename != "" {
		flags = append(flags, "--ca-file", rc.caCertFilename)
	}

	return flags
}
