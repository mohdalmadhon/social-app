package dbutil

import "strings"

var likeEscaper = strings.NewReplacer(
	`\`, `\\`,
	`%`, `\%`,
	`_`, `\_`,
)

// LikePattern turns raw user input into a "contains" pattern for queries
// written as `LIKE ? ESCAPE '\'`, so that %, _ and \ typed by the user are
// matched literally instead of acting as wildcards.
func LikePattern(search string) string {
	return "%" + likeEscaper.Replace(search) + "%"
}
