// SPDX-License-Identifier: MIT

package workspace

import (
	"strings"

	"github.com/ropeixoto/harnessx/internal/platform/constants"
)

// Slugify converts a free-form project name into the canonical slug used
// by the registry (lower-case alphanumerics separated by
// constants.SlugSeparator). Empty results fall back to
// constants.SlugFallbackName so callers never end up with an empty key.
func Slugify(in string) string {
	var b strings.Builder
	prevSeparator := false
	for _, r := range strings.ToLower(in) {
		if isSlugRune(r) {
			b.WriteRune(r)
			prevSeparator = false
			continue
		}
		if !prevSeparator {
			b.WriteString(constants.SlugSeparator)
			prevSeparator = true
		}
	}
	if out := strings.Trim(b.String(), constants.SlugSeparator); out != "" {
		return out
	}
	return constants.SlugFallbackName
}

func isSlugRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
}
