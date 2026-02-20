module github.com/Rancheroo/r8s

go 1.23.0

require (
	github.com/fatih/color v1.18.0
	github.com/spf13/cobra v1.8.1
	golang.org/x/text v0.14.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.30.0 // indirect
)

// Force golang.org/x/sys to v0.29.0 for Go 1.23 compatibility
replace golang.org/x/sys => golang.org/x/sys v0.29.0
