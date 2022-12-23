// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package create_cert

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
	"text/template"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/dir"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/filename"
	"github.com/platinasystems/goes/v2/pkg/os/host"
)

const Usage = `
usage: {{.Command}} [<options>]
Create key and certifcate for host or exchange.
{{print .Flags}}`

func Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const year = 365 * 24 * time.Hour

	defname := host.Name.Value()
	if cur, err := user.Current(); err == nil {
		switch {
		case len(cur.Name) > 0:
			defname = cur.Name
		case len(cur.Username) > 0:
			defname = cur.Username
		}
	}

	fs := flags.New()

	alg := keycert.PureEd25519
	fs.Var(&alg, "alg", strings.Join(keycert.Algs, ", "))
	sn := fs.Int64("serial-number", 1, "")
	dnsnames := fs.String("dns", host.Name.Value(), "comma separated")
	dur := fs.Duration("duration", 10*year, "note 8760 hours per year")

	email := fs.String("email", "", "")

	organization := fs.String("organization", "", "")
	locality := fs.String("locality", "", "")
	province := fs.String("province", "", "")
	country := fs.String("country", "", "")
	name := fs.String("name", defname, "")

	usage := func() error {
		return template.Must(template.New("usage").
			Parse(Usage[1:])).
			Execute(w, struct {
				Command string
				Flags   flags.Flags
			}{
				strings.Join(path, " "),
				fs,
			})
	}

	switch path[1] {
	case "complete":
		complete.Last(w, args, fs.FlagSet)
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return usage()
	}

	if err := fs.Parse(args); err == flags.ErrHelp {
		return usage()
	} else if err != nil {
		return err
	}

	if args = fs.Args(); len(args) > 0 {
		return fmt.Errorf("unexpected: %v", args)
	}

	dnsa := strings.Split(*dnsnames, ",")
	if len(dnsa) == 0 || len(dnsa[0]) == 0 {
		return errors.New("no DNS names")
	}

	k, block, err := keycert.NewPrivateKey(alg.Value())
	if err != nil {
		return err
	}
	if err = dir.Mk(); err != nil {
		return err
	}
	pemdata := pem.EncodeToMemory(block)
	err = ioutil.WriteFile(filename.PrivateKey(), pemdata, 0600)
	if err != nil {
		return err
	}
	var emails []string
	if len(*email) > 0 {
		emails = strings.Split(*email, ",")
	}
	if len(*name) == 0 {
		return errors.New("no name")
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
	if err != nil {
		return err
	}
	pemdata = pem.EncodeToMemory(block)
	return ioutil.WriteFile(filename.Cert(), pemdata, 0644)
}
