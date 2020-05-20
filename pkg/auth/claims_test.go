package auth

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScopesJSONUnmarshaling(t *testing.T) {
	claims := &claims{}
	require.NoError(t, json.Unmarshal([]byte(`{"scope": "one two three"}`), claims))
	require.Contains(t, claims.Scopes, Scope("one"))
	require.Contains(t, claims.Scopes, Scope("two"))
	require.Contains(t, claims.Scopes, Scope("three"))

	require.NoError(t, json.Unmarshal([]byte(`{"scope": "onetwothree"}`), claims))
	require.Contains(t, claims.Scopes, Scope("onetwothree"))
	require.NotContains(t, claims.Scopes, Scope("one"))
	require.NotContains(t, claims.Scopes, Scope("two"))
	require.NotContains(t, claims.Scopes, Scope("three"))

	require.Error(t, json.Unmarshal([]byte(`{"scope": 42}`), claims))
	require.Error(t, json.Unmarshal([]byte(`{"scope": true}`), claims))
	require.Error(t, json.Unmarshal([]byte(`{"scope": false}`), claims))
	require.Error(t, json.Unmarshal([]byte(`{"scope": {}}`), claims))
}
