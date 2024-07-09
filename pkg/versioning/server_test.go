package versioning

import (
	"context"
	versioning "github.com/interuss/dss/pkg/api/versioningv1"
	"github.com/interuss/dss/pkg/version"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestServer_GetVersion(t *testing.T) {
	s := &Server{}

	got := s.GetVersion(context.Background(),
		&versioning.GetVersionRequest{
			SystemIdentity: "empty",
		}).Response200

	assert.Equal(
		t,
		version.Current().String(),
		string(*got.SystemVersion),
	)

	assert.Equal(
		t,
		"empty",
		string(*got.SystemIdentity),
	)

}
