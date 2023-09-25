module github.com/platinasystems/goes/v2

go 1.21.1

require (
	github.com/creack/pty v1.1.18
	golang.org/x/net v0.12.0
	golang.org/x/sys v0.10.0
	golang.org/x/term v0.10.0
	golang/buildid v1.19.1+incompatible
)

require golang/xcoff v1.19.1+incompatible // indirect

replace (
	golang/buildid v1.19.1+incompatible => ./golang/buildid
	golang/xcoff v1.19.1+incompatible => ./golang/xcoff
)
