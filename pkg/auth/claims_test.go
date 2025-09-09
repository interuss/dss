package auth

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScopesJSONUnmarshaling(t *testing.T) {
	claims := &Claims{}
	require.NoError(t, json.Unmarshal([]byte(`{"scope": "one two three"}`), claims))
	require.Contains(t, claims.Scopes, "one")
	require.Contains(t, claims.Scopes, "two")
	require.Contains(t, claims.Scopes, "three")

	require.NoError(t, json.Unmarshal([]byte(`{"scope": "onetwothree"}`), claims))
	require.Contains(t, claims.Scopes, "onetwothree")
	require.NotContains(t, claims.Scopes, "one")
	require.NotContains(t, claims.Scopes, "two")
	require.NotContains(t, claims.Scopes, "three")

	require.Error(t, json.Unmarshal([]byte(`{"scope": 42}`), claims))
	require.Error(t, json.Unmarshal([]byte(`{"scope": true}`), claims))
	require.Error(t, json.Unmarshal([]byte(`{"scope": false}`), claims))
	require.Error(t, json.Unmarshal([]byte(`{"scope": {}}`), claims))
}
