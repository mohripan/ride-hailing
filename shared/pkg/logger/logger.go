package logger

import "go.uber.org/zap"

// New returns a production logger for "production" env, development logger otherwise.
func New(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
