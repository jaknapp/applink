package storage

import "time"

// IsExpired returns true if the token has expired
func (t *Token) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return false // No expiration set
	}
	return time.Now().After(t.ExpiresAt)
}

// NeedsRefresh returns true if the token should be refreshed
// (expires within 5 minutes)
func (t *Token) NeedsRefresh() bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().Add(5 * time.Minute).After(t.ExpiresAt)
}
