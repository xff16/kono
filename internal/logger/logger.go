package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(debug bool) *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

	var lvl zapcore.Level

	if debug {
		lvl = zap.DebugLevel
	} else {
		lvl = zap.InfoLevel
	}

	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(lvl),
		Development:      debug,
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	log, err := config.Build()
	if err != nil {
		panic(err)
	}

	return log
}
