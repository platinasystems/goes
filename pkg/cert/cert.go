// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cert

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

	"github.com/platinasystems/goes/v2/pkg/sig"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

const BlockType = "CERTIFICATE"

const (
	Year    = 365 * 24 * time.Hour
	Longest = 10 * Year
)

var (
	File string
	Flag = xflag.Label{"cert",
		"Certificate file w/in current or config directory.",
		func() any {
			var ok bool
			if File, ok = xmain.LookupEnv("CERT"); ok {
				return &File
			}
			File = "cert.pem"
			if _, err := os.Stat(File); err == nil {
			} else if _, err = os.Stat(filepath.
				Join(xmain.ConfigDir, File)); err == nil {
			} else if h, err := os.Hostname(); err == nil {
				if i := strings.Index(h, "."); i > 0 {
					h = h[:i]
				}
				File = fmt.Sprint(h, ".pem")
			}
			return &File
		}}
	Path     = func() string { return xmain.ConfigFile(File) }
	Features = map[string]any{
		"new": map[string]any{
			"certificate": New,
		},
		"show": map[string]any{
			"certificate": Show,
		},
	}
)

var NewTemplate = sync.OnceValues(func() (*template.Template, error) {
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

// Create PEM encoded x509 certificate file.
func New(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Create PEM encoded x509 certificate file.

{{flags .}}`)
	var name,
		dns,
		email,
		org,
		unit,
		street,
		city,
		state,
		country,
		zip,
		uri string
	dur := Year
	sn := int64(1)

	err := xflag.Labels{
		xmain.ConfigFlag,
		sig.Flag,
		Flag,
		{"name", "VPN identfier.", func() any {
			var err error
			if name, err = os.Hostname(); err != nil {
				return err
			} else if i := strings.Index(name, "."); i > 0 {
				name = name[:i]
			}
			return &name
		}},
		{"serial-number",
			"New certificate's identifier, random if zero.", &sn},
		{"dns", "Comma separated domain names.", func() any {
			var err error
			if dns, err = os.Hostname(); err != nil {
				return err
			}
			return &dns
		}},
		{"email", "Comma separated addresses.", &email},
		{"organization", "aka. company.", &org},
		{"organizational-unit", "aka. department.", &unit},
		{"street", `e.g. "1313 Mockingbird Lane"`, &street},
		{"locality", "aka. city.", &city},
		{"province", "aka. state.", &state},
		{"country", "Country code.", &country},
		{"postal-code", "aka. zip code.", &zip},
		{"uri", "Comma separated URLs.", &uri},
		{"duration",
			"New certificate's life span, e.g. 360s, 60m, or 1h.",
			&dur},
	}.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	if err = sig.Init(); err != nil {
		return err
	}

	if dur > Longest {
		return xerrors.Invalid(dur.String())
	}

	now := time.Now()
	expire := now.Add(dur)
	t := x509.Certificate{
		IsCA:               true,
		SerialNumber:       big.NewInt(sn),
		SignatureAlgorithm: x509.PureEd25519,
		NotBefore:          now,
		NotAfter:           expire,
		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
		Subject: pkix.Name{
			CommonName:         name,
			SerialNumber:       fmt.Sprint(sn),
			Organization:       strings.Fields(org),
			OrganizationalUnit: strings.Fields(unit),
			StreetAddress:      strings.Fields(street),
			Locality:           strings.Fields(city),
			Province:           strings.Fields(state),
			Country:            strings.Fields(country),
			PostalCode:         strings.Fields(zip),
		},
		DNSNames: strings.Split(dns, ","),
	}

	if len(email) > 0 {
		t.EmailAddresses = strings.Split(email, ",")
	}

	if len(uri) > 0 {
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
		t.NotAfter = t.NotBefore.Add(Longest)
	}

	der, err := x509.
		CreateCertificate(random, &t, parent, sig.Pub, sig.Priv)
	if err != nil {
		return err
	}
	blk := &pem.Block{
		Type:    BlockType,
		Headers: map[string]string{},
		Bytes:   der,
	}
	cp := Path()
	if cp == "-" {
		return pem.Encode(os.Stdout, blk)
	}
	if _, err = os.Stat(cp); err == nil {
		return fmt.Errorf("%s: exists", cp)
	} else if !os.IsNotExist(err) {
		return err
	}
	w, err := os.Create(cp)
	if err != nil {
		return err
	}
	defer w.Close()
	return pem.Encode(w, blk)
}

// Print parsed certificate(s).
func Show(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags]
Print parsed certificate.

{{flags .}}`)
	err := xflag.Labels{
		xmain.ConfigFlag,
		sig.Flag,
		Flag,
	}.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}

	c, err := ReadFile(Path())
	if err != nil {
		return err
	}

	t, err := NewTemplate()
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
		if !strings.HasSuffix(blk.Type, BlockType) {
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

func ReadFile(fn string) (*x509.Certificate, error) {
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
		Type:  BlockType,
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
		wc, err = xos.AppendFile(dfn, 0644)
	}
	if err != nil {
		return err
	}
	defer wc.Close()
	return pem.Encode(wc, &blk)
}

func dumpCertificates(w io.Writer, cs []*x509.Certificate) (err error) {
	blk := pem.Block{
		Type: BlockType,
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
		Type: BlockType,
	}
	fi, err := os.Stat(dfn)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return os.Remove(filepath.Join(dfn, cn+".pem"))
	}
	wc, err := os.Create(dfn)
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
