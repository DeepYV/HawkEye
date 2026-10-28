package issues

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// IssueID is the lowercase hex SHA-256 of type, route, signature, and flow name
// joined by newlines. An empty flow name is still a field, so the input ends
// with a newline.
func IssueID(issueType, route, signature, flowName string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{issueType, route, signature, flowName}, "\n")))
	return hex.EncodeToString(sum[:])
}
