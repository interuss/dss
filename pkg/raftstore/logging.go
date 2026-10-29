package raftstore

import (
	"go.uber.org/zap"
)

type ZapRaftLogger struct {
	sugar *zap.SugaredLogger
}

func (l *ZapRaftLogger) Debug(v ...interface{})                 { l.sugar.Debug(v...) }
func (l *ZapRaftLogger) Debugf(format string, v ...interface{}) { l.sugar.Debugf(format, v...) }

func (l *ZapRaftLogger) Info(v ...interface{})                 { l.sugar.Info(v...) }
func (l *ZapRaftLogger) Infof(format string, v ...interface{}) { l.sugar.Infof(format, v...) }

func (l *ZapRaftLogger) Warning(v ...interface{})                 { l.sugar.Warn(v...) }
func (l *ZapRaftLogger) Warningf(format string, v ...interface{}) { l.sugar.Warnf(format, v...) }

func (l *ZapRaftLogger) Error(v ...interface{})                 { l.sugar.Error(v...) }
func (l *ZapRaftLogger) Errorf(format string, v ...interface{}) { l.sugar.Errorf(format, v...) }

func (l *ZapRaftLogger) Panic(v ...interface{})                 { l.sugar.Panic(v...) }
func (l *ZapRaftLogger) Panicf(format string, v ...interface{}) { l.sugar.Panicf(format, v...) }

func (l *ZapRaftLogger) Fatal(v ...interface{})                 { l.sugar.Fatal(v...) }
func (l *ZapRaftLogger) Fatalf(format string, v ...interface{}) { l.sugar.Fatalf(format, v...) }
