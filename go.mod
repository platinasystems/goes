module github.com/platinasystems/goes/v2

go 1.22.0

require (
	github.com/creack/pty v1.1.18
	github.com/prometheus-community/pro-bing v0.3.0
	golang.org/x/sys v0.16.0
	golang.org/x/term v0.10.0
	golang/buildid v1.19.1+incompatible
)

require (
	github.com/google/uuid v1.3.0 // indirect
	golang.org/x/net v0.11.0 // indirect
	golang.org/x/sync v0.3.0 // indirect
	golang/xcoff v1.19.1+incompatible // indirect
)

replace (
	golang/buildid v1.19.1+incompatible => ./golang/buildid
	golang/xcoff v1.19.1+incompatible => ./golang/xcoff
)
