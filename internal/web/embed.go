// Package web bundles the companion reader into the Readlane executable.
package web
import "embed"
// Assets contains the self-hosted reader, including fonts and their licenses.
//go:embed assets
var Assets embed.FS
