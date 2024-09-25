package logger

import "go.uber.org/zap"

// var state *zap.Logger
var state *CustomLogger

type CustomLogger struct {
	*zap.Logger
}

func NewLogger() error {
	zLogger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	state = &CustomLogger{
		Logger: zLogger,
	}
	return nil
}

func Get() *CustomLogger {
	return state
}
