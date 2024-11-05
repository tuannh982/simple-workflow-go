package examples

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func GetLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return config.Build()
}
