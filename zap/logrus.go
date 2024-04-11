package logrus

import (
	"fmt"
	"io"

	"go.uber.org/zap"
)

func Debug(args ...any) {
	zap.S().Debug(args...)
}

func Info(args ...any) {
	zap.S().Info(args...)
}

func Warn(args ...any) {
	zap.S().Warn(args...)
}

func Error(args ...any) {
	zap.S().Error(args...)
}

func Panic(args ...any) {
	zap.S().Panic(args...)
}

func Fatal(args ...any) {
	zap.S().Fatal(args...)
}

func Debugf(format string, args ...any) {
	zap.S().Debugf(format, args...)
}

func Infof(format string, args ...any) {
	zap.S().Infof(format, args...)
}

func Warnf(format string, args ...any) {
	zap.S().Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	zap.S().Errorf(format, args...)
}

func Panicf(format string, args ...any) {
	zap.S().Panicf(format, args...)
}

func Fatalf(format string, args ...any) {
	zap.S().Fatalf(format, args...)
}

func Debugln(args ...any) {
	zap.S().Debugln(args...)
}

func Infoln(args ...any) {
	zap.S().Infoln(args...)
}

func Warnln(args ...any) {
	zap.S().Warnln(args...)
}

func Errorln(args ...any) {
	zap.S().Errorln(args...)
}

func Panicln(args ...any) {
	zap.S().Panicln(args...)
}

func Fatalln(args ...any) {
	zap.S().Fatalln(args...)
}

func Printf(format string, args ...any) {
	fmt.Printf(format, args...)
}

func Println(args ...any) {
	fmt.Println(args...)
}

func SetOutput(out io.Writer) {}
