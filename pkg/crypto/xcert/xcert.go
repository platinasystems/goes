// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Extended tls and x509 certificates.
package xcert

import (
	"context"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/flagctx"
	"github.com/platinasystems/goes/v2/pkg/context/pathctx"
	"github.com/platinasystems/goes/v2/pkg/context/rctx"
	"github.com/platinasystems/goes/v2/pkg/context/wctx"
	"github.com/platinasystems/goes/v2/pkg/crypto/xkey"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const GenerateUsageTemplate = `
usage: {{.Path}} [<options>],
Generate PEM encoded x509 certifcate to stdout with stdin signature key.
{{.Flag}}`

func GenerateUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		Path: pathctx.StringIn(ctx),
		Flag: flagctx.StringIn(ctx),
	}
}

func Generate(ctx context.Context, args ...string) error {
	const year = 365 * 24 * time.Hour
	const longest = 10 * year

	defname := host.Name()
	cur, err := user.Current()
	if err == nil {
		switch {
		case len(cur.Name) > 0:
			defname = cur.Name
		case len(cur.Username) > 0:
			defname = cur.Username
		}
	}

	fs := flag.NewSilentFlagSet("generate")
	ctx = flagctx.Parameter.With(ctx, fs)
	sn := fs.Int64("serial-number", 1, "")
	dnsnames := fs.String("dns", host.Name(), "comma separated list")
	dur := fs.Duration("duration", 10*year, "note 8760 hours per year")
	email := fs.String("email", "", "")
	organization := fs.String("organization", "", "")
	locality := fs.String("locality", "", "")
	province := fs.String("province", "", "")
	country := fs.String("country", "", "")
	name := fs.String("name", defname, "")

	if flag.Search[bool]("complete") {
		return complete.Last(args, fs, "*.pem")
	}
	if err = fs.Parse(args); err != nil {
		return egress.Mark(err)
	}
	if flag.Search[bool]("help", fs) {
		return usage.Error(GenerateUsageTemplate[1:],
			GenerateUsageData(ctx))

	}

	r := rctx.Parameter.In(ctx)

	var priv xkey.Private
	if _, err = priv.ReadFrom(r); err != nil {
		return egress.Mark(err)
	}

	if len(*name) == 0 {
		return egress.Mark(ErrNoName)
	}

	dnsa := strings.Split(*dnsnames, ",")
	if len(dnsa) == 0 || len(dnsa[0]) == 0 {
		return egress.Mark(ErrNoDNSNames)
	}

	var emails []string
	if len(*email) > 0 {
		emails = strings.Split(*email, ",")
	}

	now := time.Now()
	expire := now.Add(*dur)
	template := x509.Certificate{
		IsCA:               true,
		SerialNumber:       big.NewInt(*sn),
		SignatureAlgorithm: priv.SignatureAlgorithm(),
		DNSNames:           dnsa,
		EmailAddresses:     emails,
		NotBefore:          now,
		NotAfter:           expire,
		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageCertSign,
		Subject: pkix.Name{
			Organization: strings.Fields(*organization),
			Locality:     strings.Fields(*locality),
			Province:     strings.Fields(*province),
			Country:      strings.Fields(*country),
			CommonName:   *name,
		},
	}
	parent := &template

	random := rand.Reader
	if template.SerialNumber == nil {
		max := big.NewInt(math.MaxInt64)
		template.SerialNumber, err = rand.Int(random, max)
		if err != nil {
			return egress.Mark(err)
		}
	}
	if len(template.DNSNames) == 0 {
		template.DNSNames = []string{host.Name()}
	}
	if len(template.Subject.CommonName) == 0 {
		template.Subject.CommonName = template.DNSNames[0]
	}
	if template.NotBefore.IsZero() {
		template.NotBefore = time.Now()
	}
	if template.NotAfter.IsZero() ||
		template.NotAfter.Before(template.NotBefore) {
		template.NotAfter = template.NotBefore.Add(longest)
	}
	der, err := x509.CreateCertificate(random, &template, parent,
		priv.Public(), priv.Private())
	if err != nil {
		return egress.Mark(err)
	}
	blk := &pem.Block{
		Type:    "CERTIFICATE",
		Headers: map[string]string{},
		Bytes:   der,
	}
	return egress.Mark(pem.Encode(wctx.Parameter.In(ctx), blk))
}

const ShowUsageTemplate = `
usage: {{.}} [<name>]
Print decoded x509 PEM certifcate(s) from the named file or stdin.`

func ShowUsageData(ctx context.Context) any {
	return pathctx.StringIn(ctx)
}

func Show(ctx context.Context, args ...string) error {
	if flag.Search[bool]("complete") {
		return complete.Last(args, "*.pem")
	}
	if flag.Search[bool]("help") {
		return usage.Error(ShowUsageTemplate[1:], ShowUsageData(ctx))
	}
	r := rctx.Parameter.In(ctx)
	w := wctx.Parameter.In(ctx)
	if len(args) > 0 && args[0] != "-" {
		if f, err := os.Open(args[0]); err != nil {
			return err
		} else {
			defer f.Close()
			r = f
		}
	}
	var x X509
	if _, err := x.ReadFrom(r); err != nil {
		return err
	}
	for p := &x; p != nil; p = p.Next {
		fmt.Fprint(w, p)
	}
	return nil
}
