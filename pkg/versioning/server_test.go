package versioning

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/interuss/dss/pkg/api"
	versioning "github.com/interuss/dss/pkg/api/versioningv1"
	dsserr "github.com/interuss/dss/pkg/errors"
	"github.com/interuss/dss/pkg/version"
	"github.com/interuss/stacktrace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubAuthorizer struct {
	err error
}

func (a stubAuthorizer) Authorize(_ http.ResponseWriter, _ *http.Request, _ []api.AuthorizationOption) api.AuthorizationResult {
	return api.AuthorizationResult{Error: a.err}
}

func TestServer_GetVersion(t *testing.T) {
	s := &Server{}

	got := s.GetVersion(t.Context(),
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

func TestAPIRouter_SetAuthError(t *testing.T) {
	for _, tc := range []struct {
		name     string
		authErr  error
		wantCode int
		wantMsg  string
	}{
		{
			name:     "unauthenticated returns 401 with ClientErrorBody message",
			authErr:  stacktrace.NewErrorWithCode(dsserr.Unauthenticated, "missing token"),
			wantCode: http.StatusUnauthorized,
			wantMsg:  "missing token",
		},
		{
			name:     "permission denied returns 403 with ClientErrorBody message",
			authErr:  stacktrace.NewErrorWithCode(dsserr.PermissionDenied, "missing scope"),
			wantCode: http.StatusForbidden,
			wantMsg:  "missing scope",
		},
		{
			name:     "internal auth error returns 500 with InternalServerErrorBody error_message",
			authErr:  stacktrace.NewError("jwks fetch failed"),
			wantCode: http.StatusInternalServerError,
			wantMsg:  "jwks fetch failed",
		},
		{
			name:     "authorized returns 200",
			authErr:  nil,
			wantCode: http.StatusOK,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			multi := api.MultiRouter{}
			multi.SetRouter("versioning", versioning.MakeAPIRouter(&Server{}, stubAuthorizer{err: tc.authErr}))

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/versions/local_dss", nil)
			rec := httptest.NewRecorder()
			multi.ServeHTTP(rec, req)

			assert.Equal(t, tc.wantCode, rec.Code)
			switch tc.wantCode {
			case http.StatusUnauthorized, http.StatusForbidden:
				var body api.ClientErrorBody
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				require.NotNil(t, body.Message)
				assert.Contains(t, *body.Message, tc.wantMsg)
			case http.StatusInternalServerError:
				var body api.InternalServerErrorBody
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Contains(t, body.ErrorMessage, tc.wantMsg)
			}
		})
	}
}
