// Package templates embeds the built-in environment template assets.
package templates

import (
	"embed"
	"io/fs"
)

//go:embed all:*
var files embed.FS

// FS returns the embedded built-in template assets.
func FS() fs.FS { return files }
