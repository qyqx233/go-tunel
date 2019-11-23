package lib

import (
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func init() {
	log, _ := zap.NewDevelopment()
	defer log.Sync() // flushes buffer, if any
	logger = log.Sugar()
}
func GetLogger() *zap.SugaredLogger {
	return logger
}
