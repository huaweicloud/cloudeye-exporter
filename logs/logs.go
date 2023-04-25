package logs

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/cihub/seelog"
)

var Logger LoggerConstructor

type LoggerConstructor struct {
	LogInstance seelog.LoggerInterface
}

func InitLog() {
	// 注册自定义格式化控制器（防日志注入）
	registerCustomFormatter()
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

func registerCustomFormatter() {
	registerErr := seelog.RegisterCustomFormatter("CleanMsg", createCleanMsgFormatter)
	if registerErr != nil {
		log.Printf("Register msg formatter error: %s", registerErr.Error())
	}
}

func createCleanMsgFormatter(params string) seelog.FormatterFunc {
	return func(message string, level seelog.LogLevel, context seelog.LogContextInterface) interface{} {
		// 清除message中的常见日志注入字符
		message = strings.Replace(message, "\b", "", -1)
		message = strings.Replace(message, "\n", "", -1)
		message = strings.Replace(message, "\t", "", -1)
		message = strings.Replace(message, "\u000b", "", -1)
		message = strings.Replace(message, "\f", "", -1)
		message = strings.Replace(message, "\r", "", -1)
		message = strings.Replace(message, "\u007f", "", -1)
		return message
	}
}
