// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix

package vpn

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

var CertificatesTemplate = sync.OnceValues(func() (*template.Template, error) {
	return template.New("certificates").Parse(`
- dns_names:{{range .DNSNames}}
  - {{.}}{{end}}
  email_addresses:{{range .EmailAddresses}}
  - {{.}}{{end}}
  ip_addresses: {{range .IPAddresses}}
  - {{.}}{{end}}
  serial_number: {{.SerialNumber}}
  not_before: {{.NotBefore}}
  not_after: {{.NotAfter}}
  public_key_algorithm: {{.PublicKeyAlgorithm}}
  signature_algorithm: {{.SignatureAlgorithm}}
  subject:
    common_name: {{.Subject.CommonName}}
    serial_number: {{.Subject.SerialNumber}}
    organization:{{range .Subject.Organization}}
    - {{.}}{{end}}
    unit:{{range .Subject.OrganizationalUnit}}
    - {{.}}{{end}}
    street:{{range .Subject.StreetAddress}}
    - {{.}}{{end}}
    locality:{{range .Subject.Locality}}
    - {{.}}{{end}}
    province:{{range .Subject.Province}}
    - {{.}}{{end}}
    country:{{range .Subject.Country}}
    - {{.}}{{end}}
    postal_code:{{range .Subject.PostalCode}}
    - {{.}}{{end}}
    extra:{{range .Subject.ExtraNames}}
    - type: {{.Type}}
      value: {{.Value}}{{end}}
`[1:])
})

var Crt = sync.OnceValues(func() (*Certificates, error) {
	return NewCertificates(Flags.FN.Crt)
})

var Subscriptions = sync.OnceValues(func() (*Certificates, error) {
	subs, err := NewCertificates(Flags.FN.Subscriptions)
	if err == nil {
		verbose.Println(subs, "has", len(subs.BCs), "BCs")
	}
	return subs, err
})

// NewCertificate creates a PEM encoded x509 certificate file.
func NewCertificate(ctx context.Context, args []string) error {
	const year = 365 * 24 * time.Hour
	const longest = 10 * year

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
Create PEM encoded x509 certificate file.

{{flags .}}`)

	Flags.FN.Crt = filepath.Join(ConfigHome(), DefaultCrt)
	Flags.FN.Key = filepath.Join(ConfigHome(), DefaultKey)

	hostname, _ := os.Hostname()

	name := flag.String("name", hostname, "VPN identfier.")
	sn := flag.Int64("serial-number", 1, "")
	dur := flag.Duration("duration", 10*year, "e.g. 360s, 60m, or 1h.")
	dns := flag.String("dns", hostname, "Comma separated domain names.")
	email := flag.String("email", "", "Comma separated addresses.")
	organization := flag.String("organization", "", "aka. company")
	organizationalUnit := flag.String("organizational-unit", "",
		"aka. department.")
	street := flag.String("street", "", "")
	locality := flag.String("locality", "", "aka. city.")
	province := flag.String("province", "", "aka. state.")
	country := flag.String("country", "", "")
	postalCode := flag.String("postal-code", "", "aka. zip.")

	err := AddAndParseFlags(ctx, args)
	if err != nil {
		return err
	}

	key, err := Key()
	if err != nil {
		return err
	}

	priv := key.First()
	if priv == nil {
		return xerrors.Incomplete(key.String())
	}

	if *dur > longest {
		return xerrors.Invalid(key.String())
	}

	now := time.Now()
	expire := now.Add(*dur)
	t := x509.Certificate{
		IsCA:               true,
		SerialNumber:       big.NewInt(*sn),
		SignatureAlgorithm: x509.PureEd25519,
		NotBefore:          now,
		NotAfter:           expire,
		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
		Subject: pkix.Name{
			CommonName:         *name,
			SerialNumber:       fmt.Sprint(*sn),
			Organization:       strings.Fields(*organization),
			OrganizationalUnit: strings.Fields(*organizationalUnit),
			StreetAddress:      strings.Fields(*street),
			Locality:           strings.Fields(*locality),
			Province:           strings.Fields(*province),
			Country:            strings.Fields(*country),
			PostalCode:         strings.Fields(*postalCode),
		},
		DNSNames: strings.Split(*dns, ","),
	}

	if len(*email) > 0 {
		t.EmailAddresses = strings.Split(*email, ",")
	}
	parent := &t

	random := rand.Reader
	if t.SerialNumber == nil {
		max := big.NewInt(math.MaxInt64)
		t.SerialNumber, err = rand.Int(random, max)
		if err != nil {
			return xerrors.Mark(err)
		}
	}
	if t.NotBefore.IsZero() {
		t.NotBefore = time.Now()
	}
	if t.NotAfter.IsZero() || t.NotAfter.Before(t.NotBefore) {
		t.NotAfter = t.NotBefore.Add(longest)
	}

	der, err := x509.CreateCertificate(random, &t, parent,
		priv.Public(), priv)
	if err != nil {
		return err
	}
	blk := &pem.Block{
		Type:    "CERTIFICATE",
		Headers: map[string]string{},
		Bytes:   der,
	}
	if Flags.FN.Crt == "-" {
		return pem.Encode(os.Stdout, blk)
	}
	w, err := os.OpenFile(Flags.FN.Crt, oCreate, 0644)
	if err != nil {
		return err
	}
	defer w.Close()
	return pem.Encode(w, blk)
}

// Show Certificate prints parsed certificate(s).
func ShowCertificate(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
Print parsed certificate.

{{flags .}}`)

	Flags.FN.Crt = filepath.Join(ConfigHome(), DefaultCrt)

	err := AddAndParseFlags(ctx, args)
	if err != nil {
		return err
	}

	crt, err := Crt()
	if err == nil {
		err = crt.Show(os.Stdout)
	}
	return err
}

// The embedding type or method must mutex CertificatesFile.
type Certificates struct {
	// Directory or File Name
	dfn   string
	BCs   []*BC
	Named map[string]*BC
}

type BC struct {
	Block *pem.Block
	Cert  *x509.Certificate
}

// NewCertifiactes parses all of the PEM encoded [x509.Certificate](s) from
// the named directory, file, or, if named “-”, stdin.
// The returned [Certificates] retains this name to write back
// [Certificates.Add] and [Certificates.Remove].
func NewCertificates(dfn string) (*Certificates, error) {
	c := &Certificates{
		dfn:   dfn,
		Named: make(map[string]*BC),
	}
	if dfn == "-" {
		data, err := io.ReadAll(os.Stdin)
		if err == nil {
			err = c.parse(data)
		}
		return c, err
	} else if strings.HasSuffix(dfn, ".pem") {
		data, err := os.ReadFile(dfn)
		if err == nil {
			if err = c.parse(data); err == nil {
				verbose.Println("parsed:", dfn)
			}
		}
		return c, err
	}
	fns, err := filepath.Glob(filepath.Join(dfn,
		filepath.FromSlash("/*.pem")))
	if err != nil {
		return c, err
	}
	for _, fn := range fns {
		if data, err := os.ReadFile(fn); err != nil {
			return c, err
		} else if err = c.parse(data); err != nil {
			return c, xerrors.Label(err, fn)
		}
		verbose.Println("parsed:", fn)
	}
	return c, nil
}

// Decode PEM then parse certificate from DER.
func (cs *Certificates) parse(data []byte) error {
	var blks []*pem.Block
	for blk, r := pem.Decode(data); blk != nil; blk, r = pem.Decode(r) {
		if strings.HasSuffix(blk.Type, "CERTIFICATE") {
			blks = append(blks, blk)
		}
	}
	for i, blk := range blks {
		b := blk.Bytes
		if x, err := x509.ParseCertificate(b); err != nil {
			return xerrors.Label(err, "block", fmt.Sprint(i))
		} else if x != nil {
			bc := &BC{blk, x}
			cs.BCs = append(cs.BCs, bc)
			cs.Named[x.Subject.CommonName] = bc
		}
	}
	return nil
}

func (cs *Certificates) Add(blk *pem.Block, x *x509.Certificate) error {
	cn := x.Subject.CommonName
	if _, ok := cs.Named[cn]; ok {
		return xerrors.Unavailable(cn)
	}
	bc := &BC{blk, x}
	cs.BCs = append(cs.BCs, bc)
	cs.Named[cn] = bc
	if cs.dfn == "-" {
		return pem.Encode(os.Stdout, blk)
	}
	if strings.HasSuffix(cs.dfn, ".pem") {
		w, err := os.OpenFile(cs.dfn, oAppend, 0644)
		if err != nil {
			return err
		}
		defer w.Close()
		return pem.Encode(w, blk)
	}
	f, err := os.CreateTemp(cs.dfn, "*.pem")
	if err == nil {
		defer f.Close()
		err = pem.Encode(f, blk)
	}
	return err
}

func (cs *Certificates) DERs() [][]byte {
	ders := make([][]byte, len(cs.BCs))
	for i, bc := range cs.BCs {
		ders[i] = bc.Block.Bytes
	}
	return ders
}

func (cs *Certificates) Dump(w io.Writer) error {
	for _, bc := range cs.BCs {
		if err := pem.Encode(w, bc.Block); err != nil {
			return err
		}
	}
	return nil
}

// Thie returns <nil> if there are no certificates.
func (cs *Certificates) First() *x509.Certificate {
	for _, bc := range cs.BCs {
		if bc.Cert != nil {
			return bc.Cert
		}
	}
	return nil
}

func (cs *Certificates) Has(peer *x509.Certificate) bool {
	bc, found := cs.Named[peer.Subject.CommonName]
	if !found {
		return false
	}
	return peer.Equal(bc.Cert)
}

func (cs *Certificates) Join(pool *x509.CertPool) {
	for _, bc := range cs.BCs {
		if bc.Cert != nil {
			verbose.Println("+root:", bc.Cert.Subject.CommonName)
			pool.AddCert(bc.Cert)
		}
	}
}

func (cs *Certificates) Lookup(cn string) (
	*pem.Block, *x509.Certificate, error,
) {
	bc, ok := cs.Named[cn]
	if !ok {
		return nil, nil, xerrors.NotFound(cn)
	}
	return bc.Block, bc.Cert, nil
}

func (cs *Certificates) Remove(cn string) error {
	bc, ok := cs.Named[cn]
	if !ok {
		return xerrors.NotFound(cn)
	}
	bc.Block = nil
	bc.Cert = nil
	delete(cs.Named, cn)
	if cs.dfn == "-" {
		return nil
	}
	if strings.HasSuffix(cs.dfn, ".pem") {
		w, err := os.OpenFile(cs.dfn, oCreate, 0644)
		if err != nil {
			return err
		}
		defer w.Close()
		for _, bc := range cs.BCs {
			if bc.Block != nil {
				if err = pem.Encode(w, bc.Block); err != nil {
					return err
				}
			}
		}
		return nil
	}
	dir, err := os.ReadDir(cs.dfn)
	if err != nil {
		return err
	}
	for _, de := range dir {
		if !strings.HasSuffix(de.Name(), ".pem") {
			break
		}
		fn := filepath.Join(cs.dfn, de.Name())
		if err != nil {
			continue
		}
		data, err := os.ReadFile(fn)
		if err != nil {
			continue
		}
		for blk, r := pem.Decode(data); blk != nil; blk, r = pem.
			Decode(r) {
			if !strings.HasSuffix(blk.Type, "CERTIFICATE") {
				continue
			}
			x, err := x509.ParseCertificate(blk.Bytes)
			if err == nil && x != nil &&
				x.Subject.CommonName == cn {
				return os.Remove(fn)
			}

		}
	}
	return nil
}

func (cs *Certificates) Show(w io.Writer) error {
	if len(cs.BCs) == 0 {
		fmt.Println(w, "# none")
	} else if t, err := CertificatesTemplate(); err != nil {
		return err
	} else {
		for _, bc := range cs.BCs {
			if err = t.Execute(w, bc.Cert); err != nil {
				return err
			}
		}
	}
	return nil
}

func (cs *Certificates) String() string {
	return cs.dfn
}
