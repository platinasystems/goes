// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
)

const BlockTypeCertificate = "CERTIFICATE"

const (
	year    = 365 * 24 * time.Hour
	longest = 10 * year
)

var CertFile = xflag.New[string]("cert", `
Certificate file w/in current or config directory.
`[1:], func() string {
	s, ok := xmain.LookupEnv("CERT")
	if ok {
		return s
	}
	s = "cert.pem"
	if _, err := os.Stat(s); err == nil {
	} else if _, err = os.Stat(filepath.
		Join(xmain.Config.Value(), s)); err == nil {
	} else if h, err := os.Hostname(); err == nil {
		if i := strings.Index(h, "."); i > 0 {
			h = h[:i]
		}
		s = fmt.Sprint(h, ".pem")
	}
	return s
})

func CertPath() string {
	return xmain.Config.File(CertFile.Value())
}

var CertDuration = xflag.New[time.Duration]("duration", `
New certificate's life span, e.g. 360s, 60m, or 1h.
`[1:], func() time.Duration {
	return year
})

var CertSerialNumber = xflag.New[int64]("serial-number", `
New certificate's identifier, random if zero.
`[1:], func() int64 {
	return 1
})

var CertCountry = xflag.New[string]("country", `
New certificate's country code.`[1:], nil)

var CertDNS = xflag.New[string]("dns", `
Comma separated domain names.
`[1:], func() string {
	s, err := os.Hostname()
	if err != nil {
		s = err.Error()
	}
	return s
})

var CertEmail = xflag.New[string]("email", `
New certificate's comma separated addresses.`[1:], nil)

var CertLocality = xflag.New[string]("locality", `
aka. city.`[1:], nil)

var CertName = xflag.New[string]("name", `
VPN identfier.
`[1:], func() string {
	s, err := os.Hostname()
	if err != nil {
		s = err.Error()
	} else if i := strings.Index(s, "."); i > 0 {
		s = s[:i]
	}
	return s
})

var CertOrganization = xflag.New[string]("organization", `
aka. company.`[1:], nil)

var CertOrganizationalUnit = xflag.New[string]("organizational-unit", `
aka. department.`[1:], nil)

var CertPostalCode = xflag.New[string]("postal-code", `
aka. zip code.`[1:], nil)

var CertProvince = xflag.New[string]("province", `
aka. state.`[1:], nil)

var CertStreet = xflag.New[string]("street", `
e.g. "1313 Mockingbird Lane"`[1:], nil)

var CertURI = xflag.New[string]("uri", `
Comma separated URLs.`[1:], nil)

var CertificatesTemplate = sync.OnceValues(func() (*template.Template, error) {
	return template.New("certificates").Parse(`{{range .}}- {{/*
*/}}dns_names:{{range .DNSNames}}
  - {{.}}{{end}}
  email_addresses:{{range .EmailAddresses}}
  - {{.}}{{/*
*/}}{{end}}{{/*
*/}}{{if .IPAddresses}}
  ip_addresses:{{range .IPAddresses}}
  - {{.}}{{/*
*/}}{{end}}{{end}}{{/*
*/}}{{if .URIs}}
  uris:{{range .URIs}}
  - {{.}}{{/*
*/}}{{end}}{{end}}
  serial_number: {{.SerialNumber}}
  not_before: {{.NotBefore}}
  not_after: {{.NotAfter}}
  public_key_algorithm: {{.PublicKeyAlgorithm}}
  signature_algorithm: {{.SignatureAlgorithm}}
  subject:
    common_name: {{.Subject.CommonName}}
    serial_number: {{.Subject.SerialNumber}}{{/*
*/}}{{if .Subject.Organization}}
    organization:{{range .Subject.Organization}}
    - {{.}}{{/*
*/}}{{end}}{{end}}{{/*
*/}}{{if  .Subject.OrganizationalUnit}}
    unit:{{range .Subject.OrganizationalUnit}}
    - {{.}}{{/*
*/}}{{end}}{{end}}{{/*
*/}}{{if .Subject.StreetAddress}}
    street:{{range .Subject.StreetAddress}}
    - {{.}}{{/*
*/}}{{end}}{{end}}{{/*
*/}}{{if .Subject.Locality}}
    locality:{{range .Subject.Locality}}
    - {{.}}{{/*
*/}}{{end}}{{end}}{{/*
*/}}{{if .Subject.Province}}
    province:{{range .Subject.Province}}
    - {{.}}{{/*
*/}}{{end}}{{end}}{{/*
*/}}{{if .Subject.Country}}
    country:{{range .Subject.Country}}
    - {{.}}{{/*
*/}}{{end}}{{end}}{{/*
*/}}{{if .Subject.PostalCode}}
    postal_code:{{range .Subject.PostalCode}}
    - {{.}}{{/*
*/}}}{{end}}{{end}}{{/*
*/}}{{if .Subject.ExtraNames}}
    extra:{{range .Subject.ExtraNames}}
    - type: {{.Type}}
      value: {{.Value}}{{/*
*/}}{{end}}{{end}}
{{end}}`)
})

// CreateCertificate a PEM encoded x509 certificate file.
func CreateCertificate(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Create PEM encoded x509 certificate file.

{{flags .}}`)

	xmain.Config.Define()

	CertFile.Define()
	CertCountry.Define()
	CertDNS.Define()
	CertDuration.Define()
	CertEmail.Define()
	CertLocality.Define()
	CertName.Define()
	CertOrganization.Define()
	CertOrganizationalUnit.Define()
	CertPostalCode.Define()
	CertProvince.Define()
	CertSerialNumber.Define()
	SigFile.Define()
	CertStreet.Define()
	CertURI.Define()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	if err = signInit(); err != nil {
		return err
	}

	if CertDuration.Value() > longest {
		return xerrors.Invalid(CertDuration.Value().String())
	}

	now := time.Now()
	expire := now.Add(CertDuration.Value())
	t := x509.Certificate{
		IsCA:               true,
		SerialNumber:       big.NewInt(CertSerialNumber.Value()),
		SignatureAlgorithm: x509.PureEd25519,
		NotBefore:          now,
		NotAfter:           expire,
		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
		Subject: pkix.Name{
			CommonName:   CertName.String(),
			SerialNumber: CertSerialNumber.String(),
			Organization: strings.Fields(CertOrganization.String()),
			OrganizationalUnit: strings.
				Fields(CertOrganizationalUnit.String()),
			StreetAddress: strings.Fields(CertStreet.String()),
			Locality:      strings.Fields(CertLocality.String()),
			Province:      strings.Fields(CertProvince.String()),
			Country:       strings.Fields(CertCountry.String()),
			PostalCode:    strings.Fields(CertPostalCode.String()),
		},
		DNSNames: strings.Split(CertDNS.Value(), ","),
	}

	if s := CertEmail.Value(); len(s) > 0 {
		t.EmailAddresses = strings.Split(s, ",")
	}

	if uri := CertURI.Value(); len(uri) > 0 {
		for _, s := range strings.Split(uri, ",") {
			u, err := url.Parse(s)
			if err != nil {
				return err
			}
			t.URIs = append(t.URIs, u)
		}
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
		signPub, signPriv)
	if err != nil {
		return err
	}
	blk := &pem.Block{
		Type:    BlockTypeCertificate,
		Headers: map[string]string{},
		Bytes:   der,
	}
	cp := CertPath()
	if cp == "-" {
		return pem.Encode(os.Stdout, blk)
	}
	if _, err = os.Stat(cp); err == nil {
		return fmt.Errorf("%s: exists", cp)
	} else if !os.IsNotExist(err) {
		return err
	}
	w, err := os.OpenFile(cp, oCreate, 0644)
	if err != nil {
		return err
	}
	defer w.Close()
	return pem.Encode(w, blk)
}

// Print parsed certificate(s).
func ShowCertificate(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Print parsed certificate.

{{flags .}}`)

	xmain.Config.Define()
	CertFile.Define()
	SigFile.Define()

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	c, err := readCertificateFile(CertPath())
	if err != nil {
		return err
	}

	t, err := CertificatesTemplate()
	if err != nil {
		return err
	}
	return t.Execute(os.Stdout, []*x509.Certificate{c})
}

// This parses all of the PEM encoded [x509.Certificate](s) from
// the named directory, file, or, if named “-”, stdin.
func certificates(dfn string) (cs []*x509.Certificate, err error) {
	var fns []string
	var rc io.ReadCloser
	var cś []*x509.Certificate
	if dfn == "-" {
		cs, err = readCertificates(os.Stdin)
		return
	}
	fi, err := os.Stat(dfn)
	if err != nil {
		return
	}
	if fi.IsDir() {
		fns, err = filepath.Glob(filepath.Join(dfn, "*.pem"))
		if err != nil {
			return
		}
	} else {
		fns = append(fns, dfn)
	}
	for _, fn := range fns {
		if rc, err = os.Open(fn); err != nil {
			return
		} else {
			cś, err = readCertificates(rc)
			rc.Close()
			if err != nil {
				return
			}
			cs = append(cs, cś...)
		}
	}
	return
}

func readCertificates(r io.Reader) (cs []*x509.Certificate, err error) {
	var eof bool

	ŕ := bufio.NewReaderSize(r, 64<<10)
	b := make([]byte, 0, 4<<10)

	for {
		if !eof {
			n, erŕ := ŕ.Read(b[len(b):cap(b)])
			if n > 0 {
				b = b[:len(b)+n]
			} else {
				eof = errors.Is(erŕ, io.EOF)
				if !eof {
					err = erŕ
					break
				}
			}
		}
		blk, rem := pem.Decode(b)
		if blk == nil {
			break
		}
		b = b[:len(rem)]
		copy(b, rem)
		if !strings.HasSuffix(blk.Type, BlockTypeCertificate) {
			continue
		}
		c, erŕ := x509.ParseCertificate(blk.Bytes)
		if erŕ != nil {
			err = erŕ
			break
		}
		cs = append(cs, c)
	}
	return
}

func readCertificateFile(fn string) (*x509.Certificate, error) {
	b, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	blk, _ := pem.Decode(b)
	if blk == nil {
		return nil, xerrors.Invalid(fn)
	}
	return x509.ParseCertificate(blk.Bytes)
}

func addCertificate(dfn string, c *x509.Certificate) error {
	var wc io.WriteCloser
	blk := pem.Block{
		Type:  BlockTypeCertificate,
		Bytes: c.Raw,
	}
	cn := c.Subject.CommonName
	fi, err := os.Stat(dfn)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		wc, err = os.Create(filepath.Join(dfn, cn+".pem"))
	} else {
		wc, err = os.OpenFile(dfn, oAppend, 0644)
	}
	if err != nil {
		return err
	}
	defer wc.Close()
	return pem.Encode(wc, &blk)
}

func dumpCertificates(w io.Writer, cs []*x509.Certificate) (err error) {
	blk := pem.Block{
		Type: BlockTypeCertificate,
	}
	for _, c := range cs {
		blk.Bytes = c.Raw
		if err = pem.Encode(w, &blk); err != nil {
			return err
		}
	}
	return nil
}

func removeCertificate(dfn, cn string, cs []*x509.Certificate) error {
	blk := pem.Block{
		Type: BlockTypeCertificate,
	}
	fi, err := os.Stat(dfn)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return os.Remove(filepath.Join(dfn, cn+".pem"))
	}
	wc, err := os.OpenFile(dfn, oCreate, 0644)
	if err != nil {
		return err
	}
	defer wc.Close()
	for _, c := range cs {
		if c.Subject.CommonName == cn {
			continue
		}
		blk.Bytes = c.Raw
		if err = pem.Encode(wc, &blk); err != nil {
			return err
		}
	}
	return nil
}
