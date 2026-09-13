package cmd

import _ "embed"

//go:embed docs/root.txt
var rootLong string

//go:embed docs/init.txt
var initLong string

//go:embed docs/update.txt
var updateLong string

//go:embed docs/version.txt
var versionLong string
