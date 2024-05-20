// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/crypto/x509/x509certs"
)

type pending struct {
	sync.RWMutex
	blocks []*pem.Block
	certs  []*x509.Certificate
}

func (p *pending) add(block *pem.Block, cert *x509.Certificate) {
	p.Lock()
	defer p.Unlock()
	for i, c := range p.certs {
		if c == nil {
			p.blocks[i] = block
			p.certs[i] = cert
			return
		}
	}
	p.blocks = append(p.blocks, block)
	p.certs = append(p.certs, cert)
}

func (p *pending) pull(cn string) (*pem.Block, *x509.Certificate, error) {
	p.Lock()
	defer p.Unlock()
	for i, cert := range p.certs {
		if cert.Subject.CommonName == cn {
			block := p.blocks[i]
			p.blocks[i] = nil
			p.certs[i] = nil
			return block, cert, nil
		}
	}
	return nil, nil, fmt.Errorf("%s: %w", cn, ErrNotFound)
}

func (p *pending) Format(w fmt.State, verb rune) {
	p.RLock()
	defer p.RUnlock()
	if len(p.certs) == 0 {
		fmt.Fprintln(w, "# none")
	} else if t, err := x509certs.Template(); err != nil {
		fmt.Fprintln(w, err)
	} else {
		t.Execute(w, p.certs)
	}
}
