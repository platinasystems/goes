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
	"math"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xpem"
)

func newX509Certificate(ctx context.Context, args []string) error {
	const year = 365 * 24 * time.Hour
	const longest = 10 * year

	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
Generate PEM encoded x509 certificate.

{{flags .}}`)

	opts.crt = flag.String("o", defaultCrt(),
		"Output file name or “-” for stdout.")
	opts.key = flag.String("k", defaultKey(),
		"Key file name or “-” for stdin.")

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

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	k, err := keyFile()
	if err != nil {
		return err
	}

	priv := k.First()
	if priv == nil {
		return xerrors.Incomplete(k.Path)
	}

	if *dur > longest {
		return xerrors.Invalid(k.Path)
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
	block := &pem.Block{
		Type:    "CERTIFICATE",
		Headers: map[string]string{},
		Bytes:   der,
	}
	if *opts.crt == "-" {
		err = xpem.EncodeAll(os.Stdout, block)
	} else {
		err = xpem.Create(*opts.crt, 0644, block)
	}
	return err
}

func showX509Certificate(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags]
Print parsed certificate.

{{flags .}}`)

	opts.crt = flag.String("i", defaultCrt(),
		"Input file name or “-” for stdin.")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}

	crt, err := crtFile()
	if err == nil {
		err = crt.Show(os.Stdout)
	}
	return err
}
