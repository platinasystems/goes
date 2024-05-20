// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix

package vpn

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/pem/pemblocks"
	"github.com/platinasystems/goes/v2/pkg/crypto/x509/x509keys"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

var generate = map[string]any{
	"certificate": generateCertificate,
	"key":         generateKey,
}

func generateKey(ctx context.Context, args []string) error {
	const year = 365 * 24 * time.Hour
	const longest = 10 * year
	const usage = `
usage: {{branch .}} [<options>] [<output>]
Generate PEM encoded ed25519 key.

The default output is "{{key}}";
use '-' for stdout.`

	k, _ := vpnKeyFile()
	if goes.ContextComplete(ctx) {
		return nil
	}
	if goes.ContextHelp(ctx) {
		goes.TemplateFuncs["key"] = vpnKeyPath
		return goes.Usage(ctx, usage)
	}
	if len(args) > 0 {
		k.Path = args[0]
	}
	_, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return err
	}
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	block := &pem.Block{
		Type:    "PRIVATE KEY",
		Headers: map[string]string{},
		Bytes:   der,
	}
	if k.Path == "-" {
		err = pemblocks.Encode(goes.ContextStdout(ctx), block)
	} else {
		err = pemblocks.Create(k.Path, 0600, block)
	}
	return err
}

func generateCertificate(ctx context.Context, args []string) error {
	const year = 365 * 24 * time.Hour
	const longest = 10 * year
	const usage = `
usage: {{branch .}} [<options>] [<output> [<key>]]
Generate PEM encoded x509 certificate.

The default <output> is "{{certificate}}";
use '-' for stdout.

The default <key> is "{{key}}";
use '-' for stdin.
{{flags .}}`
	var flags flag.FlagSet
	ctx = goes.FlagsContext(ctx, &flags)
	name := flags.String("name", host.Name(), "VPN identfier.")
	sn := flags.Int64("serial-number", 1, "")
	dur := flags.Duration("duration", 10*year, "e.g. 360s, 60m, or 1h.")
	dns := flags.String("dns", host.Name(), "Comma separated domain names.")
	email := flags.String("email", "", "Comma separated addresses.")
	organization := flags.String("organization", "", "aka. company")
	organizationalUnit := flags.String("organizational-unit", "",
		"aka. department.")
	street := flags.String("street", "", "")
	locality := flags.String("locality", "", "aka. city.")
	province := flags.String("province", "", "aka. state.")
	country := flags.String("country", "", "")
	postalCode := flags.String("postal-code", "", "aka. zip.")
	if goes.ContextComplete(ctx) {
		if len(args) > 1 {
			return complete.Last(args, "*.pem")
		}
		return nil
	}
	ctx, err := goes.ParseFlagsContext(ctx, args)
	if err != nil {
		return err
	}
	if goes.ContextHelp(ctx) {
		goes.TemplateFuncs["key"] = vpnKeyPath
		goes.TemplateFuncs["certificate"] = vpnCrtPath
		return goes.Usage(ctx, usage)
	}

	args = flags.Args()

	path := vpnCrtPath()
	if len(args) > 0 {
		path = args[0]
	}

	var k *x509keys.File
	if len(args) > 1 {
		if args[1] == "-" {
			k = &x509keys.File{Path: "-"}
			_, err = k.ReadFrom(goes.ContextStdin(ctx))
		} else {
			k, err = x509keys.NewFile(args[1])
		}
	} else {
		k, err = vpnKeyFile()
	}
	if err != nil {
		return err
	}

	priv := k.First()
	if priv == nil {
		return fmt.Errorf("%s: %w", k.Path, ErrNoPrivateKeys)
	}

	if *dur > longest {
		return ErrTooLong
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
			return egress.Mark(err)
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
	if path == "-" {
		err = pemblocks.Encode(goes.ContextStdout(ctx), block)
	} else {
		err = pemblocks.Create(path, 0644, block)
	}
	return err
}
