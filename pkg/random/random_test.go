package random

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratorIsDeterministicForTheSameSeedAndLabel(t *testing.T) {
	generator1, err := Generator(42, "label")
	require.NoError(t, err)
	generator2, err := Generator(42, "label")
	require.NoError(t, err)

	var buf1, buf2 [16]byte
	_, err = generator1.Read(buf1[:])
	require.NoError(t, err)
	_, err = generator2.Read(buf2[:])
	require.NoError(t, err)

	require.Equal(t, buf1, buf2)
}

func TestGeneratorDiffersForDifferentSeeds(t *testing.T) {
	generator1, err := Generator(42, "label")
	require.NoError(t, err)
	generator2, err := Generator(43, "label")
	require.NoError(t, err)

	var buf1, buf2 [16]byte
	_, err = generator1.Read(buf1[:])
	require.NoError(t, err)
	_, err = generator2.Read(buf2[:])
	require.NoError(t, err)

	require.NotEqual(t, buf1, buf2)
}

func TestGeneratorDiffersForDifferentLabels(t *testing.T) {
	generator1, err := Generator(42, "label-a")
	require.NoError(t, err)
	generator2, err := Generator(42, "label-b")
	require.NoError(t, err)

	var buf1, buf2 [16]byte
	_, err = generator1.Read(buf1[:])
	require.NoError(t, err)
	_, err = generator2.Read(buf2[:])
	require.NoError(t, err)

	require.NotEqual(t, buf1, buf2)
}
