// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math"
	"math/big"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

var DefaultSelfFileName = sync.OnceValue(func() string {
	return filepath.Join(xdg.StateHome(), program.Base(),
		"tlsx", "self.pem")
})

var OptionalSelfFileName *string

var Self struct {
	X509
	tls.Certificate
}

var InitSelf = sync.OnceValue(func() error {
	sigdata, err := os.ReadFile(*OptionalSignatureFileName)
	if err != nil {
		return err
	}
	if err = Self.X509.UnmarshalText(sigdata); err != nil {
		return nil
	}
	certdata, err := os.ReadFile(*OptionalSelfFileName)
	if err != nil {
		return nil
	}
	err = Self.X509.UnmarshalText(certdata)
	if err == nil {
		err = ValidateSelf()
	}
	return err
})

func GenerateSelf(ctx context.Context, args []string) error {
	const year = 365 * 24 * time.Hour
	const longest = 10 * year

	var flags flag.FlagSet
	ctx = goes.FlagsContext(ctx, &flags)

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
	sn := flags.Int64("serial-number", 1, "")
	dnsnames := flags.String("dns", host.Name(), "comma separated list")
	dur := flags.Duration("duration", 10*year, "note 8760 hours per year")
	email := flags.String("email", "", "")
	organization := flags.String("organization", "", "")
	locality := flags.String("locality", "", "")
	province := flags.String("province", "", "")
	country := flags.String("country", "", "")
	name := flags.String("name", defname, "")

	if goes.ContextComplete(ctx) {
		return complete.Last(args, &flags, "*.pem")
	}
	ctx, err = goes.ParseFlagsContext(ctx, args)
	if err != nil {
		return err
	}
	if goes.ContextHelp(ctx) {
		goes.TemplateFuncs["filename"] = *OptionalSelfFileName
		return goes.Usage(ctx, `
usage: {{branch .}} [<options>]
Generate PEM encoded {{filename}}.
{{flags .}}`)
	}
	args = flags.Args()

	err = InitSelf()
	if err != nil {
		return err
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
		SignatureAlgorithm: x509.PureEd25519,
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
		Signature.Public, Signature.Private)
	if err != nil {
		return egress.Mark(err)
	}
	w, err := goes.ContextOutput(ctx, *OptionalSelfFileName, 0600)
	if err != nil {
		return egress.Mark(err)
	}
	return pem.Encode(w, &pem.Block{
		Type:    "CERTIFICATE",
		Headers: map[string]string{},
		Bytes:   der,
	})
}

func ShowSelf(ctx context.Context, args []string) error {
	if goes.ContextComplete(ctx) {
		return complete.Last(args, "*.pem")
	}
	if goes.ContextHelp(ctx) {
		goes.TemplateFuncs["filename"] = func() string {
			return *OptionalSelfFileName
		}
		return goes.Usage(ctx, `
usage: {{branch .}}
Print x509 certifcate within PEM encoded {{filename}}.`)
	}
	err := InitSelf()
	if err != nil {
		return err
	}
	w := goes.ContextStdout(ctx)
	algs := Self.Certificate.SupportedSignatureAlgorithms
	if n := len(algs); n > 0 {
		fmt.Fprintln(w, "supported_signature_algoritums:")
		for _, alg := range algs {
			fmt.Fprintln(w, "  -", alg)
		}
	}
	ctss := Self.Certificate.SignedCertificateTimestamps
	if n := len(ctss); n > 0 {
		fmt.Fprintln(w, "signed_certificate_timestamps:", n)
	}
	fmt.Fprint(w, &Self.X509)
	return nil
}

func ValidateSelf() error {
	if Self.Certificate.PrivateKey == nil {
		return egress.Mark(ErrPrivateKey)
	}
	if len(Self.X509.Certificate.DNSNames) == 0 {
		return egress.Mark(ErrNoDNSNames)
	}
	if Self.X509.Block.Type != "CERTIFICATE" {
		return egress.Mark(ErrInvalid)
	}
	pub, ok := Self.X509.Certificate.PublicKey.(ed25519.PublicKey)
	if !ok {
		return egress.Mark(ErrInvalid)
	}
	priv, ok := Self.Certificate.PrivateKey.(ed25519.PrivateKey)
	if !ok {
		return egress.Mark(ErrKeyType)
	}
	if !bytes.Equal(priv.Public().(ed25519.PublicKey), pub) {
		return egress.Mark(ErrKeyParm)
	}
	Self.Certificate.Certificate =
		append(Self.Certificate.Certificate, Self.X509.Block.Bytes)
	for next := Self.X509.Next; next != nil; next = Self.Next {
		Self.Certificate.Certificate =
			append(Self.Certificate.Certificate, next.Block.Bytes)
	}
	return nil
}
