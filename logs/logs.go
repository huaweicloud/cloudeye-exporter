package logs

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/cihub/seelog"
)

var Logger LoggerConstructor

type LoggerConstructor struct {
	LogInstance seelog.LoggerInterface
}

func InitLog() {
	data, err := ioutil.ReadFile("./logs.conf")
	if err != nil {
		log.Fatalf("Failed to read log config file, error: %s", err.Error())
	}

	logger, err := seelog.LoggerFromConfigAsBytes(data)
	if err != nil {
		log.Fatalf("Failed to init log, error: %s", err.Error())
	}
	Logger.LogInstance = logger
}

func (lc *LoggerConstructor) Flush() {
	lc.LogInstance.Flush()
}

func (lc *LoggerConstructor) Tracef(format string, params ...interface{}) {
	lc.LogInstance.Tracef(format, params...)
}

func (lc *LoggerConstructor) Debugf(format string, params ...interface{}) {
	lc.LogInstance.Debugf(format, params...)
}

func (lc *LoggerConstructor) Infof(format string, params ...interface{}) {
	lc.LogInstance.Infof(format, params...)
}

func (lc *LoggerConstructor) Warnf(format string, params ...interface{}) {
	lc.LogInstance.Warnf(format, params...) // nolint:errcheck
}

func (lc *LoggerConstructor) Errorf(format string, params ...interface{}) {
	lc.LogInstance.Errorf(format, params...) // nolint:errcheck
}

func (lc *LoggerConstructor) Criticalf(format string, params ...interface{}) {
	lc.LogInstance.Criticalf(format, params...) // nolint:errcheck
}

func (lc *LoggerConstructor) Trace(params ...interface{}) {
	lc.LogInstance.Trace(params...)
}

func (lc *LoggerConstructor) Debug(params ...interface{}) {
	lc.LogInstance.Debug(params...)
}

func (lc *LoggerConstructor) Info(params ...interface{}) {
	lc.LogInstance.Info(params...)
}

func (lc *LoggerConstructor) Warn(params ...interface{}) {
	lc.LogInstance.Warn(params...) // nolint:errcheck
}

func (lc *LoggerConstructor) Error(params ...interface{}) {
	lc.LogInstance.Error(params...) // nolint:errcheck
}

func (lc *LoggerConstructor) Critical(params ...interface{}) {
	lc.LogInstance.Critical(params...) // nolint:errcheck
}

func FlushLogAndExit(code int) {
	Logger.Flush()
	os.Exit(code)
}
