// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os/user"
	"strings"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/filename"
	"github.com/platinasystems/goes/v2/pkg/os/host"
)

func CreateCert(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	const year = 365 * 24 * time.Hour

	hn, err := host.Name.ValErr()
	if err != nil {
		return err
	}

	defname := hn
	if cur, err := user.Current(); err == nil {
		switch {
		case len(cur.Name) > 0:
			defname = cur.Name
		case len(cur.Username) > 0:
			defname = cur.Username
		}
	}

	keyfn := filename.PrivateKey.String()
	certfn := filename.Cert.String()
	fs := flag.NewFlagSet("create", flag.ContinueOnError)
	alg := keycert.PureEd25519
	fs.Var(&alg, "alg", fmt.Sprint(keycert.Algs))
	sn := fs.Int64("serial-number", 1, "")
	dnsnames := fs.String("dns", hn, "comma separated")
	dur := fs.Duration("duration", 10*year, "note 8760 hours per year")

	email := fs.String("email", "", "")

	organization := fs.String("organization", "", "")
	locality := fs.String("locality", "", "")
	province := fs.String("province", "", "")
	country := fs.String("country", "", "")
	name := fs.String("name", defname, "")

	fs.Usage = func() {
		path.Usage(w, "[<options>]\n",
			"Create key and certifcate for"+
				" host, consumer, or exchange.\n",
			fs,
		)
	}

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() > 0 {
		return fmt.Errorf("unexpected: %v", fs.Args())
	}

	if path.HasComplete() {
		complete.Last(w, args, fs, keycert.Algs)
		return nil
	}
	if path.HasHelp() {
		fs.Usage()
		return nil
	}

	dnsa := strings.Split(*dnsnames, ",")
	if len(dnsa) == 0 || len(dnsa[0]) == 0 {
		return errors.New("no DNS names")
	}

	k, block, err := keycert.NewPrivateKey(alg.Value())
	if err != nil {
		return err
	}
	if err = state.MkDir(); err != nil {
		return err
	}
	pemdata := pem.EncodeToMemory(block)
	err = ioutil.WriteFile(keyfn, pemdata, 0600)
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
	return ioutil.WriteFile(certfn, pemdata, 0644)
}
