// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package httpx

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
)

func NewClientCfg(
	server string,
	certs []tls.Certificate,
	roots ...any,
) (*tls.Config, error) {
	cfg := new(tls.Config)
	cfg.ServerName = server
	cfg.Certificates = certs
	cfg.InsecureSkipVerify = true
	return cfg, AddRootCAs(cfg, roots...)
}

func NewServerCfg(name string, certs []tls.Certificate, clients ...any) (
	*tls.Config, error,
) {
	cfg := new(tls.Config)
	cfg.ServerName = name
	cfg.Certificates = certs
	cfg.ClientAuth = tls.RequestClientCert
	return cfg, AddClientCAs(cfg, clients...)
}

// types of clients:
//	*x509.Certificate
//	[]byte, PEM data
//	string, PEM encoded file
func AddClientCAs(cfg *tls.Config, clients ...any) error {
	if len(clients) == 0 {
		return nil
	}
	if cfg.ClientCAs == nil {
		cfg.ClientCAs = x509.NewCertPool()
	}
	cfg.ClientAuth = tls.RequireAndVerifyClientCert
	for _, v := range clients {
		switch t := v.(type) {
		case *x509.Certificate:
			cfg.ClientCAs.AddCert(t)
		case []byte:
			cfg.ClientCAs.AppendCertsFromPEM(t)
		case string:
			pem, err := ioutil.ReadFile(t)
			if err != nil {
				return err
			}
			cfg.ClientCAs.AppendCertsFromPEM(pem)
		}
	}
	return nil
}

// types of roots:
//	*x509.Certificate
//	[]byte, PEM data
//	string, PEM encoded file
func AddRootCAs(cfg *tls.Config, roots ...any) error {
	if len(roots) == 0 {
		return nil
	}
	if cfg.RootCAs == nil {
		cfg.RootCAs = x509.NewCertPool()
	}
	if cfg.InsecureSkipVerify {
		cfg.InsecureSkipVerify = false
	}
	for _, v := range roots {
		switch t := v.(type) {
		case *x509.Certificate:
			cfg.RootCAs.AddCert(t)
		case []byte:
			cfg.RootCAs.AppendCertsFromPEM(t)
		case string:
			pem, err := ioutil.ReadFile(t)
			if err != nil {
				return err
			}
			cfg.RootCAs.AppendCertsFromPEM(pem)
		}
	}
	return nil
}
