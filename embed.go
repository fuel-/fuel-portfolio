// Package portfolio embeds the site's templates and static assets so the
// build produces a single self-contained binary.
package portfolio

import "embed"

//go:embed templates
var Templates embed.FS

//go:embed static
var Static embed.FS
