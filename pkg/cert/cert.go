// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cert

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"math"
	"math/big"
	"net/http"
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
const Ext = ".pem"
const CertFile = "cert.pem"

const (
	Year    = 365 * 24 * time.Hour
	Longest = 10 * Year
)

var ClientCommonName = sync.OnceValue(func() string {
	s, err := os.Hostname()
	if err != nil {
		s = "localhost"
	} else if i := strings.Index(s, "."); i > 0 {
		s = s[:i]
	}
	return s
})

var (
	Client      string
	ClientCerts = sync.OnceValues(func() ([]*x509.Certificate, error) {
		if len(Client) == 0 {
			env := fmt.Sprint("$", xmain.EnvPrefix(), "CLIENT_CN")
			return nil, xerrors.Incomplete(env)
		}
		return namedCerts(Client)
	})
	ClientFlag = xflag.Label{"ssl-client-cn", `
The file or common name (CN) of the client certificate.
(or $<main>_CLIENT_CN, $SSL_CLIENT_CN)`[1:], func() *string {
		if s, ok := xmain.LookupEnv("CLIENT_CN"); ok {
			Client = s
		} else if s, ok = os.LookupEnv("SSL_CLIENT_CN"); ok {
			Client = s
		} else {
			Client = ClientCommonName()
		}
		return &Client
	}}
)

var MainConfigAndStateDirCerts = sync.OnceValues(func() (
	certs []*x509.Certificate, err error,
) {
	certs, err = MainConfigDirCerts()
	if err == nil {
		var state []*x509.Certificate
		if state, err = MainStateDirCerts(); err == nil {
			certs = append(certs, state...)
		}
	}
	return
})

// Parse *.pem files w/in [xmain.ConfigDir] for [BlockType] certificates.
var MainConfigDirCerts = sync.OnceValues(func() (
	[]*x509.Certificate, error,
) {
	return parse(xmain.ConfigDir)
})

// [MainConfigCerts] added to [x509.SystemCertPool].
func MainConfigDirPlusSystemCertPool() (*x509.CertPool, error) {
	pool, err := x509.SystemCertPool()
	if err == nil {
		var cs []*x509.Certificate
		if cs, err = MainConfigDirCerts(); err == nil {
			for _, c := range cs {
				pool.AddCert(c)
			}
		}
	}
	return pool, err
}

// If [xmain.StateDir] doesn't equal [xmain.ConfigDir],
// parse all of its *.pem files for [BlockType] certificates.
var MainStateDirCerts = sync.OnceValues(func() (
	state []*x509.Certificate, err error,
) {
	if xmain.StateDir != xmain.ConfigDir {
		state, err = parse(xmain.StateDir)
	}
	return
})

var (
	Server      string
	ServerCerts = sync.OnceValues(func() ([]*x509.Certificate, error) {
		return namedCerts(Server)
	})
	ServerFlag = xflag.Label{"ssl-server-dn", `
The file or distinguished name (DN) of the server certificate.
(or $<main>_SERVER_DN, $SSL_SERVER_DN)`[1:], func() *string {
		if s, ok := xmain.LookupEnv("SERVER_DN"); ok {
			Server = s
		} else if s, ok = os.LookupEnv("SSL_SERVER_DN"); ok {
			Server = s
		} else {
			Server = Client
		}
		return &Server
	}}
)

var (
	Verify     = true
	VerifyFlag = xflag.Label{"ssl-verify", `
Verify server certificate.
(or $SSL_VERIFY)`[1:], &Verify}
)

var Features = map[string]any{
	"new": map[string]any{
		"certificate": New,
	},
	"show": map[string]any{
		"certificate": Show,
	},
}

var NewTemplate = sync.OnceValues(func() (*template.Template, error) {
	return template.New("certs").Parse(`{{range .}}- {{/*
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

const NewCertUsage = `
usage: {{.Name}} [flags] [- | <filename>]
Create PEM encoded x509 certificate file.
The default filename is “` + CertFile + `” w/in the “-config” directory.

{{flags .}}`

// Create PEM encoded x509 certificate file.
func New(ctx context.Context, args []string) error {
	xflag.TemplateUsage(NewCertUsage)
	var dns, email, org, unit, street, city, state, country, zip,
		uri string
	dur := Year
	sn := int64(1)

	err := xflag.Labels{
		xmain.ConfigFlag,
		sig.Flag,
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
	args = flag.CommandLine.Args()

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
			CommonName:         ClientCommonName(),
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
	var fn string
	if len(args) > 0 {
		fn = args[0]
	} else {
		fn = xmain.ConfigFile(CertFile)
	}
	if fn == "-" {
		return pem.Encode(os.Stdout, blk)
	} else if _, err = os.Stat(fn); err == nil {
		return fmt.Errorf("%s: exists", fn)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	w, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer w.Close()
	return pem.Encode(w, blk)
}

func NewTransport() (*http.Transport, error) {
	tp := http.DefaultTransport.(*http.Transport).Clone()
	rootCAs, err := MainConfigDirPlusSystemCertPool()
	tp.TLSClientConfig = &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: !Verify,
		RootCAs:            rootCAs,
	}
	return tp, err
}

// Print parsed certificate(s).
func Show(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] [-|<file>|<dir>]...
Print parsed certificate.

{{flags .}}`)
	err := xflag.Labels{
		xmain.ConfigFlag,
		xmain.StateFlag,
		sig.Flag,
		ClientFlag,
	}.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	}
	args = flag.CommandLine.Args()

	t, err := NewTemplate()
	if err != nil {
		return err
	}

	var match []*x509.Certificate
	if len(args) == 0 {
		if match, err = ClientCerts(); err != nil {
			return err
		}
	} else {
		for _, arg := range args {
			if cs, err := namedCerts(arg); err != nil {
				return err
			} else {
				match = append(match, cs...)
			}
		}
	}
	return t.Execute(os.Stdout, match)
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

func decode(r io.Reader) ([]*x509.Certificate, error) {
	var cs []*x509.Certificate
	var eof bool

	br := bufio.NewReaderSize(r, 64<<10)
	b := make([]byte, 0, 4<<10)

	for {
		if !eof {
			n, err := br.Read(b[len(b):cap(b)])
			if n > 0 {
				b = b[:len(b)+n]
			} else if eof = errors.Is(err, io.EOF); !eof {
				return cs, err
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
		if c, err := x509.ParseCertificate(blk.Bytes); err != nil {
			return cs, err
		} else {
			cs = append(cs, c)
		}
	}
	return cs, nil
}

func dumpCerts(w io.Writer, cs []*x509.Certificate) (err error) {
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

func namedCerts(name string) ([]*x509.Certificate, error) {
	var match []*x509.Certificate
	if strings.HasSuffix(name, Ext) {
		if r, err := tryOpen(name); err != nil {
			return nil, err
		} else {
			defer r.Close()
			return decode(r)
		}
	}
	certs, err := MainConfigAndStateDirCerts()
	if err == nil {
		for _, c := range certs {
			if c.Subject.CommonName == name {
				match = append(match, c)
			}
		}
		if len(match) == 0 {
			err = xerrors.NotFound(name)
		}
	}
	return match, err
}

// Parse all of the PEM encoded [x509.Certificate](s) within
// dir, file, or, if named “-”, stdin.
func parse(dfn string) ([]*x509.Certificate, error) {
	var fns []string
	var cs []*x509.Certificate
	if dfn == "-" {
		return xerrors.MarkResult(decode(os.Stdin))
	}
	if strings.HasSuffix(dfn, Ext) {
		fns = append(fns, dfn)
	} else if _, err := os.Stat(dfn); err != nil {
		// ignore missing directory
		return nil, nil
	} else {
		pat := filepath.Join(dfn, "*"+Ext)
		if fns, err = filepath.Glob(pat); err != nil {
			return nil, xerrors.Mark(err)
		}
	}
	for _, fn := range fns {
		if rc, err := os.Open(fn); err != nil {
			return cs, err
		} else {
			more, err := decode(rc)
			rc.Close()
			if err != nil {
				return cs, xerrors.Mark(err)
			}
			cs = append(cs, more...)
		}
	}
	return cs, nil
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

func tryOpen(name string) (io.ReadCloser, error) {
	r, err := os.Open(name)
	if err == nil {
		return r, err
	} else if !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	return os.Open(filepath.Join(xmain.ConfigDir, name))
}
