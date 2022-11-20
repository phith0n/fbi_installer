package html

import "embed"

//go:embed assets/* favicon.ico index.html
var StaticFS embed.FS
