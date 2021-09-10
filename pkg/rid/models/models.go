package models

import (
	"net/url"
	"github.com/interuss/stacktrace"
)

// ValidateUSSBaseURL ensures https
func ValidateUSSBaseURL(s string) error {
	u, err := url.Parse(s)
	if err != nil {
		return stacktrace.Propagate(err, "Error parsing URL")
	}

	switch u.Scheme {
	case "https":
		// All good, proceed normally.
	case "http":
		return stacktrace.NewError("uss_base_url in new_subscription must use TLS")
	default:
		return stacktrace.NewError("uss_base_url must support https scheme")
	}

	return nil
}
