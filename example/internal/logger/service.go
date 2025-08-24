package logger

import "go.uber.org/zap"

// ProvideLogger creates a new zap logger for development
func ProvideLogger() (*zap.Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	return logger, nil
}
