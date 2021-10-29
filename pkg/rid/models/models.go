package models

import (
	"github.com/interuss/stacktrace"
	"net/url"
)

// ValidateURL ensures https
func ValidateURL(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return stacktrace.Propagate(err, "Error parsing URL")
	}

	switch u.Scheme {
	case "https":
		// All good, proceed normally.
	case "http":
		return stacktrace.NewError("rid url must use TLS")
	default:
		return stacktrace.NewError("rid url must support https scheme")
	}

	return nil
}
