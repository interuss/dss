package logging

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc/grpclog"
)

type zapGrpcLogger struct {
	logger    *zap.Logger
	verbosity int
}

func grpcReplaceLogger(logger *zap.Logger) {
	zgl := &zapGrpcLogger{
		logger:    logger,
		verbosity: 0,
	}
	grpclog.SetLoggerV2(zgl)
}

func (l *zapGrpcLogger) Info(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Infoln(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Infof(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLogger) Warning(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Warningln(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Warningf(format string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Errorln(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Fatalln(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLogger) V(level int) bool {
	return level <= l.verbosity
}
