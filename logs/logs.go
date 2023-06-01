package logs

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v3"
)

var Logger LoggerConstructor

type LoggerConstructor struct {
	LogInstance logger
}

// logger 接口
type logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
	Sync() error
}

// Config Zap日志配置项
type Config struct {
	Level      zapcore.Level `yaml:"level"`
	Type       string        `yaml:"type"`
	Filename   string        `yaml:"filename,omitempty"`
	Encoder    string        `yaml:"encoder,omitempty"`
	TimeFormat string        `yaml:"time_format,omitempty"`
	MaxSize    int           `yaml:"max_size,omitempty"` // byte
	MaxBackups int           `yaml:"max_backups,omitempty"`
	MaxAge     int           `yaml:"max_age,omitempty"`
	Enabled    bool          `yaml:"enabled"`
	Compress   bool          `yaml:"compress"`
}

type Unmarshal func(data []byte, cfg interface{}) error

type ConfLoader struct {
	Unmarshal Unmarshal
}

func newYamlLoader() *ConfLoader {
	return newConfLoader(yaml.Unmarshal)
}

func newConfLoader(u Unmarshal) *ConfLoader {
	return &ConfLoader{
		Unmarshal: u,
	}
}

func (c ConfLoader) LoadFile(fPath string, cfg interface{}) error {
	data, err := ioutil.ReadFile(fPath)
	if err != nil {
		return err
	}
	return c.LoadData(data, cfg)
}

func (c ConfLoader) LoadData(data []byte, cfg interface{}) error {
	err := c.Unmarshal(data, cfg)
	return err
}

func InitLog() {
	var cfg map[string][]Config
	err := newYamlLoader().LoadFile("./logs.yml", &cfg)
	if err != nil {
		fmt.Printf("Fail to load logs.yml, error: %s", err.Error())
		return
	}
	config, ok := cfg["business"]
	if !ok {
		fmt.Printf("logs.yml should contains business config.")
		return
	}

	Logger.LogInstance = makeZapLogger(config).WithOptions(zap.AddCallerSkip(1)).Sugar()
}

func (zap *LoggerConstructor) Debug(args ...interface{}) {
	zap.LogInstance.Debug(clearLineBreaks("", args...))
}
func (zap *LoggerConstructor) Info(args ...interface{}) {
	zap.LogInstance.Info(clearLineBreaks("", args...))
}
func (zap *LoggerConstructor) Warn(args ...interface{}) {
	zap.LogInstance.Warn(clearLineBreaks("", args...))
}
func (zap *LoggerConstructor) Error(args ...interface{}) {
	zap.LogInstance.Error(clearLineBreaks("", args...))
}
func (zap *LoggerConstructor) Fatal(args ...interface{}) {
	zap.LogInstance.Fatal(clearLineBreaks("", args...))
}
func (zap *LoggerConstructor) Debugf(template string, args ...interface{}) {
	zap.LogInstance.Debugf(clearLineBreaks(template, args...))
}
func (zap *LoggerConstructor) Infof(template string, args ...interface{}) {
	zap.LogInstance.Infof(clearLineBreaks(template, args...))
}
func (zap *LoggerConstructor) Warnf(template string, args ...interface{}) {
	zap.LogInstance.Warnf(clearLineBreaks(template, args...))
}
func (zap *LoggerConstructor) Errorf(template string, args ...interface{}) {
	zap.LogInstance.Errorf(clearLineBreaks(template, args...))
}
func (zap *LoggerConstructor) Fatalf(template string, args ...interface{}) {
	zap.LogInstance.Fatalf(clearLineBreaks(template, args...))
}
func (zap *LoggerConstructor) Flush() {
	err := zap.LogInstance.Sync()
	if err != nil {
		fmt.Printf("Fail to sync logs, error: %s", err.Error())
	}
}

func FlushLogAndExit(code int) {
	Logger.Flush()
	os.Exit(code)
}

// zap取message方法
func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}
	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}
	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

func clearLineBreaks(template string, args ...interface{}) string {
	message := getMessage(template, args)
	if message != "" {
		// 清除message中的常见日志注入字符
		message = strings.Replace(message, "\b", "", -1)
		message = strings.Replace(message, "\n", "", -1)
		message = strings.Replace(message, "\t", "", -1)
		message = strings.Replace(message, "\u000b", "", -1)
		message = strings.Replace(message, "\f", "", -1)
		message = strings.Replace(message, "\r", "", -1)
		message = strings.Replace(message, "\u007f", "", -1)
	}
	return message
}

func makeRotate(file string, maxSize int, maxBackups int, maxAge int, compress bool) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   file,
		MaxSize:    maxSize, // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAge, // days
		LocalTime:  true,
		Compress:   compress,
	}
}

func makeEncoder(c *Config) zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(ts time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(ts.Format(c.TimeFormat))
	}
	encoderConfig.EncodeDuration = func(d time.Duration, encoder zapcore.PrimitiveArrayEncoder) {
		val := float64(d) / float64(time.Millisecond)
		encoder.AppendString(fmt.Sprintf("%.3fms", val))
	}
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	if c.Encoder == "JSON" {
		return zapcore.NewJSONEncoder(encoderConfig)
	}
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func makeWriteSync(c *Config) zapcore.WriteSyncer {
	if c.Type == "FILE" {
		logRotate := makeRotate(c.Filename, c.MaxSize/1024/1024, c.MaxBackups, c.MaxAge, c.Compress)
		return zapcore.AddSync(logRotate)
	}
	return zapcore.AddSync(os.Stdout)
}

func makeZapCore(c *Config) zapcore.Core {
	encoder := makeEncoder(c)
	w := makeWriteSync(c)
	core := zapcore.NewCore(encoder, w, c.Level)
	return core
}

func makeZapLogger(cfg []Config) *zap.Logger {
	cores := make([]zapcore.Core, 0, len(cfg))
	for i := range cfg {
		if !cfg[i].Enabled {
			continue
		}
		cores = append(cores, makeZapCore(&cfg[i]))
	}
	return zap.New(zapcore.NewTee(cores...), zap.AddCaller())
}
