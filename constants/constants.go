package constants

import _ "embed"

//go:embed git_commit.txt
var GIT_COMMIT string

//go:embed version.txt
var VERSION string
