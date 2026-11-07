// Package seeddata embeds the versioned seed content files.
package seeddata

import "embed"

// Files contains the versioned seed content loaded on first start.
//
//go:embed *.json
var Files embed.FS
