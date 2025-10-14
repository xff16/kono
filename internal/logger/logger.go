package logger

import "go.uber.org/zap"

func Init(debug bool) *zap.Logger {
	var log *zap.Logger
	var err error
	if debug {
		log, err = zap.NewDevelopment()
	} else {
		log, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}

	return log
}
