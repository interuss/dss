package raftstore

import (
	"go.uber.org/zap"
)

type ZapRaftLogger struct {
	sugar *zap.SugaredLogger
}

func (l *ZapRaftLogger) Debug(v ...any)                 { l.sugar.Debug(v...) }
func (l *ZapRaftLogger) Debugf(format string, v ...any) { l.sugar.Debugf(format, v...) }

func (l *ZapRaftLogger) Info(v ...any)                 { l.sugar.Info(v...) }
func (l *ZapRaftLogger) Infof(format string, v ...any) { l.sugar.Infof(format, v...) }

func (l *ZapRaftLogger) Warning(v ...any)                 { l.sugar.Warn(v...) }
func (l *ZapRaftLogger) Warningf(format string, v ...any) { l.sugar.Warnf(format, v...) }

func (l *ZapRaftLogger) Error(v ...any)                 { l.sugar.Error(v...) }
func (l *ZapRaftLogger) Errorf(format string, v ...any) { l.sugar.Errorf(format, v...) }

func (l *ZapRaftLogger) Panic(v ...any)                 { l.sugar.Panic(v...) }
func (l *ZapRaftLogger) Panicf(format string, v ...any) { l.sugar.Panicf(format, v...) }

func (l *ZapRaftLogger) Fatal(v ...any)                 { l.sugar.Fatal(v...) }
func (l *ZapRaftLogger) Fatalf(format string, v ...any) { l.sugar.Fatalf(format, v...) }
