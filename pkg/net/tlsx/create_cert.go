// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os/user"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

func CreateCert(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<options>]
Create key and certifcate for host or exchange.
{{print .Flags}}`
	const year = 365 * 24 * time.Hour

	defname := host.Name()
	if cur, err := user.Current(); err == nil {
		switch {
		case len(cur.Name) > 0:
			defname = cur.Name
		case len(cur.Username) > 0:
			defname = cur.Username
		}
	}

	fs, h := flag.New()
	alg := keycert.PureEd25519
	fs.Var(&alg, "alg", strings.Join(keycert.Algs, ", "))
	sn := fs.Int64("serial-number", 1, "")
	dnsnames := fs.String("dns", host.Name(), "comma separated")
	dur := fs.Duration("duration", 10*year, "note 8760 hours per year")
	email := fs.String("email", "", "")
	organization := fs.String("organization", "", "")
	locality := fs.String("locality", "", "")
	province := fs.String("province", "", "")
	country := fs.String("country", "", "")
	name := fs.String("name", defname, "")
	if complete.Parameter.Value(ctx) {
		style.Completions(args, fs.FlagSet)
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if help.Parameter.Value(ctx) || *h {
		return style.Usage(usage, struct {
			Path  []string
			Flags fmt.Formatter
		}{path, fs})
	}
	args = fs.Args()
	if len(args) > 0 {
		return fmt.Errorf("unexpected: %v", args)
	}
	if len(*name) == 0 {
		return errors.New("no name")
	}

	dnsa := strings.Split(*dnsnames, ",")
	if len(dnsa) == 0 || len(dnsa[0]) == 0 {
		return errors.New("no DNS names")
	}

	var emails []string
	if len(*email) > 0 {
		emails = strings.Split(*email, ",")
	}

	if err = MkStateDir(); err != nil {
		return egress.Marked(err)
	}

	k, block, err := keycert.NewPrivateKey(alg.Value())
	if err != nil {
		return egress.Marked(err)
	}

	pemdata := pem.EncodeToMemory(block)
	err = ioutil.WriteFile(KeyFileName(), pemdata, 0600)
	if err != nil {
		return egress.Marked(err)
	}

	now := time.Now()
	expire := now.Add(*dur)
	template := x509.Certificate{
		IsCA:               true,
		SerialNumber:       big.NewInt(*sn),
		SignatureAlgorithm: alg.Value(),
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
	_, block, err = keycert.NewX509Certificate(k, &template)
	if egress.Marked(err); err != nil {
		return err
	}
	pemdata = pem.EncodeToMemory(block)
	return ioutil.WriteFile(CertFileName(), pemdata, 0644)
}
