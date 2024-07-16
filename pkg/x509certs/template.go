// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package x509certs

import (
	"sync"
	"text/template"
)

var Template = sync.OnceValues(func() (*template.Template, error) {
	return template.New("certificates").Parse(`{{/*
*/}}{{ $n := len .}}{{if eq $n 0}}# none
{{else}}{{range .}}{{if .}}- dns_names:{{range .DNSNames}}
  - {{.}}{{end}}
  email_addresses:{{range .EmailAddresses}}
  - {{.}}{{end}}
  ip_addresses: {{range .IPAddresses}}
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
{{end}}{{end}}{{end}}`)
})
