package emails

import _ "embed"

//go:embed verification.html
var VERIFICATION string

//go:embed reset_password.html
var RESET_PASSWORD string
