package models

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	restapi "github.com/interuss/dss/pkg/api/scdv1"
	dssmodels "github.com/interuss/dss/pkg/models"
	"github.com/interuss/stacktrace"
)

const (
	// Value for OVN that should be returned for entities not owned by the client
	NoOvnPhrase = "Available from USS"

	// NullV4UUID is a valid V4 UUID to be used as a placeholder where no UUID is available
	// but one needs to be specified, as in certain API return values.
	// Note that this UUID is not meant to be persisted to the database: it should only be used
	// to populate required API fields for which a proper value does not exist.
	NullV4UUID = restapi.SubscriptionID("00000000-0000-4000-8000-000000000000")

	// maxClockSkew is the largest allowed interval between a client-provided
	// time and the server's idea of the current time.
	maxClockSkew = time.Minute * 5
)

type (
	// OVN models an opaque version number.
	OVN string

	// Version models the version of an entity.
	// Primarily used as a fencing token in data mutations.
	VersionNumber int32
)

// NewOVNFromTime encodes t as an OVN.
func NewOVNFromTime(t time.Time, salt string) OVN {
	sum := sha256.Sum256([]byte(salt + t.Format(time.RFC3339Nano)))
	ovn := base64.StdEncoding.EncodeToString(
		sum[:],
	)
	ovn = strings.ReplaceAll(ovn, "+", "-")
	ovn = strings.ReplaceAll(ovn, "/", ".")
	ovn = strings.ReplaceAll(ovn, "=", "_")
	return OVN(ovn)
}

// NewOVNFromUUIDv7Suffix returns an OVN based on an UUIDv7 suffix: `{op_intent_id}_{uuidv7_suffix}`.
// It validates that the suffix is indeed a UUIDv7 and that its timestamp is not too far from now.
func NewOVNFromUUIDv7Suffix(now time.Time, oiID dssmodels.ID, suffix string) (OVN, error) {
	uuidV7, err := uuid.Parse(suffix)
	if err != nil {
		return "", stacktrace.Propagate(err, "Suffix `%s` is not a valid UUID", suffix)
	}
	if uuidV7.Version() != 7 {
		return "", stacktrace.NewError("Suffix `%s` is not version 7 but version %d", suffix, uuidV7.Version())
	}

	var (
		ovnTime = time.Unix(uuidV7.Time().UnixTime())
		skew    = now.Sub(ovnTime).Abs()
	)
	if skew > maxClockSkew {
		return "", stacktrace.NewError("Suffix `%s` is too far away from now (got %s, max is %s)", suffix, skew.String(), maxClockSkew.String())
	}

	return OVN(fmt.Sprintf("%s_%s", oiID.String(), suffix)), nil
}

// Empty returns true if ovn indicates an empty opaque version number.
func (ovn OVN) Empty() bool {
	return len(ovn) == 0
}

// Valid returns true if ovn is valid.
func (ovn OVN) Valid() bool {
	return len(ovn) >= 16 && len(ovn) <= 128
}

func (ovn OVN) String() string {
	return string(ovn)
}

// Empty returns true if the value of v indicates an empty version.
func (v VersionNumber) Empty() bool {
	return v <= 0
}

// Matches returns true if v matches w.
func (v VersionNumber) Matches(w VersionNumber) bool {
	return v == w
}

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
		return stacktrace.NewError("uss_base_url must use TLS")
	default:
		return stacktrace.NewError("uss_base_url must support https scheme")
	}

	return nil
}
