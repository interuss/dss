package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func NewObserver() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.InfoLevel)
	return zap.New(core), logs
}

func RequireNoLogging(t *testing.T, logs *observer.ObservedLogs) {
	require.Equal(t, len(logs.All()), 0)
}

func RequireDeprecationWarning(t *testing.T, logs *observer.ObservedLogs) {
	entry := logs.All()[0]
	require.Equal(t, entry.Level, zap.WarnLevel)
	require.Equal(t, entry.Message, "DEPRECATED: enable_http has been renamed to allow_http_base_urls.")
}

func TestHttpDeprecationDoesNothingWhenBothFlagsAreOff(t *testing.T) {
	logger, logs := NewObserver()
	newFlag := new(bool)
	oldFlag := new(bool)
	SetDeprecatingHttpFlag(logger, &newFlag, &oldFlag)
	RequireNoLogging(t, logs)
	require.False(t, newFlag == oldFlag)
	require.False(t, *newFlag)
	require.False(t, *oldFlag)
}

func TestHttpDeprecationDoesNothingWhenNewFlagIsOn(t *testing.T) {
	logger, logs := NewObserver()
	newFlag := new(bool)
	oldFlag := new(bool)
	*newFlag = true
	SetDeprecatingHttpFlag(logger, &newFlag, &oldFlag)
	RequireNoLogging(t, logs)
	require.False(t, newFlag == oldFlag)
	require.True(t, *newFlag)
	require.False(t, *oldFlag)
}

func TestHttpDeprecationReassignsAddressWhenOldFlagIsOn(t *testing.T) {
	logger, logs := NewObserver()
	newFlag := new(bool)
	oldFlag := new(bool)
	*oldFlag = true
	SetDeprecatingHttpFlag(logger, &newFlag, &oldFlag)
	RequireDeprecationWarning(t, logs)
	require.True(t, newFlag == oldFlag)
	require.True(t, *newFlag)
	require.True(t, *oldFlag)
}

func TestHttpDeprecationPrefersNewFlagWhenBothAreOn(t *testing.T) {
	logger, logs := NewObserver()
	newFlag := new(bool)
	oldFlag := new(bool)
	*newFlag = true
	*oldFlag = true
	SetDeprecatingHttpFlag(logger, &newFlag, &oldFlag)
	RequireDeprecationWarning(t, logs)
	require.False(t, newFlag == oldFlag)
	require.True(t, *newFlag)
	require.True(t, *oldFlag)
}
