// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

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
	"net/url"
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
	return template.New("certificates").Parse(`{{range .}}
- dns_names:{{range .DNSNames}}
  - {{.}}{{end}}
  email_addresses:{{range .EmailAddresses}}
  - {{.}}{{end}}
  ip_addresses: {{range .IPAddresses}}
  - {{.}}{{end}}
  uris:{{range .URIs}}
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
{{end}}`[1:])
})

// CreateCertificate a PEM encoded x509 certificate file.
func CreateCertificate(ctx context.Context, args []string) error {
	const year = 365 * 24 * time.Hour
	const longest = 10 * year

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
Create PEM encoded x509 certificate file.

{{flags .}}`)

	cf := xflag.LastName(flag.CommandLine)
	kflag := KeyFlag()
	dfn := filepath.Join(ConfigDir(), fmt.Sprint(cf, ".pem"))
	oflag := flag.String("o", dfn, "Output file name, “-” for stdout.")

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
	uris := flag.String("uri", "", "Comma separated URLs.")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	sig, err := NewSignatures(*kflag)
	if err != nil {
		return err
	}

	priv := sig.First()
	if priv == nil {
		return xerrors.Incomplete(sig.String())
	}

	if *dur > longest {
		return xerrors.Invalid(dur.String())
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

	if len(*uris) > 0 {
		for _, s := range strings.Split(*uris, ",") {
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
		priv.Public(), priv)
	if err != nil {
		return err
	}
	blk := &pem.Block{
		Type:    "CERTIFICATE",
		Headers: map[string]string{},
		Bytes:   der,
	}
	if *oflag == "-" {
		return pem.Encode(os.Stdout, blk)
	}
	if _, err = os.Stat(*oflag); err == nil {
		return fmt.Errorf("%s: exists", *oflag)
	} else if !os.IsNotExist(err) {
		return err
	}
	w, err := os.OpenFile(*oflag, oCreate, 0644)
	if err != nil {
		return err
	}
	defer w.Close()
	return pem.Encode(w, blk)
}

// Print parsed certificate(s).
func ShowCertificate(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
Print parsed certificate.

{{flags .}}`)

	cf := xflag.LastName(flag.CommandLine)
	dfn := filepath.Join(ConfigDir(), fmt.Sprint(cf, ".pem"))
	iflag := flag.String("i", dfn,
		"X509 certificate file name, “-” for stdin.")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	cs, err := certificates(*iflag)
	if err != nil {
		return err
	}
	if len(cs) == 0 {
		fmt.Println("# none")
		return nil
	}
	t, err := CertificatesTemplate()
	if err != nil {
		return err
	}
	return t.Execute(os.Stdout, cs)
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

// FIXME replace os.ReadAll+pem.Decode with something that scans stream.
func readCertificates(r io.Reader) (cs []*x509.Certificate, err error) {
	var c *x509.Certificate
	data, err := io.ReadAll(r)
	if err != nil {
		return
	}
	for blk, r := pem.Decode(data); blk != nil; blk, r = pem.Decode(r) {
		if !strings.HasSuffix(blk.Type, "CERTIFICATE") {
			continue
		}
		if c, err = x509.ParseCertificate(blk.Bytes); err != nil {
			return
		} else {
			cs = append(cs, c)
		}
	}
	return
}

func addCertificate(dfn string, c *x509.Certificate) error {
	var wc io.WriteCloser
	blk := pem.Block{
		Type:  "CERTIFICATE",
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
		Type: "CERTIFICATE",
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
		Type: "CERTIFICATE",
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
