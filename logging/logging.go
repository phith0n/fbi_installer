package logging

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger zap.Logger
var sugar zap.SugaredLogger

func GetLogger() *zap.Logger {
	return &logger
}

func GetSugar() *zap.SugaredLogger {
	return &sugar
}

func InitLogger(debug bool) error {
	if t, err := NewZapLogger(debug); err != nil {
		return nil
	} else {
		logger = *t
		sugar = *logger.Sugar()
	}
	return nil
}

func NewZapLogger(debug bool) (*zap.Logger, error) {
	var config zap.Config

	if debug {
		config = zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.DebugLevel),
			Development:      true,
			Encoding:         "console",
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
	} else {
		config = zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:      false,
			Encoding:         "console",
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
	}

	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:          "T",
		LevelKey:         "L",
		NameKey:          "N",
		CallerKey:        zapcore.OmitKey,
		FunctionKey:      zapcore.OmitKey,
		MessageKey:       "M",
		StacktraceKey:    zapcore.OmitKey,
		ConsoleSeparator: " ",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalColorLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		},
		EncodeName: func(loggerName string, enc zapcore.PrimitiveArrayEncoder) {
			// Print logger name in cyan (ANSI code 36).
			enc.AppendString(fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(36), "["+loggerName+"]"))
		},
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	l, err := config.Build()
	if err != nil {
		return nil, err
	} else {
		return l, nil
	}
}
