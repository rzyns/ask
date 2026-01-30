// Package web provides the static web assets for the application.
package web

import "embed"

// Assets contains the embedded web frontend.
//
//go:embed all:index.html css js img
var Assets embed.FS
