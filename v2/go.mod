module github.com/platinasystems/goes/v2

go 1.19

require (
	github.com/creack/pty v1.1.18
	golang.org/x/sys v0.0.0-20220926163933-8cfa568d3c25
	golang.org/x/term v0.0.0-20220919170432-7a66f970e087
	golang/buildid v1.19.1+incompatible
)

require golang/xcoff v1.19.1+incompatible // indirect

replace (
	golang/buildid v1.19.1+incompatible => ./golang/buildid
	golang/xcoff v1.19.1+incompatible => ./golang/xcoff
)
