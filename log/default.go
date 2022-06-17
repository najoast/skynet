package log

import (
	"fmt"
	"time"
)

type DefaultLogger struct {
	Prefix string
	Level  Level
}

func strLevel(level Level) string {
	switch level {
	case Debug:
		return "[DEBUG] "
	case Info:
		return "[INFO ] "
	case Warn:
		return "[WARN ] "
	case Error:
		return "[ERROR] "
	case Fatal:
		return "[FATAL] "
	}
	return ""
}

func strTime() string {
	return time.Now().Format("2006-01-02 15:04:05.999 ")
}

func (logger *DefaultLogger) Outputf(level Level, format string, v ...interface{}) {
	if level >= logger.Level {
		fmt.Println(strTime() + strLevel(level) + logger.Prefix + fmt.Sprintf(format, v...))
	}
}

func (logger *DefaultLogger) Output(level Level, v ...interface{}) {
	if level >= logger.Level {
		fmt.Println(strTime() + strLevel(level) + logger.Prefix + fmt.Sprint(v...))
	}
}

func (logger *DefaultLogger) Debug(v ...interface{}) {
	logger.Output(Debug, v...)
}

func (logger *DefaultLogger) Debugf(format string, v ...interface{}) {
	logger.Outputf(Debug, format, v...)
}

func (logger *DefaultLogger) Info(v ...interface{}) {
	logger.Output(Info, v...)
}

func (logger *DefaultLogger) Infof(format string, v ...interface{}) {
	logger.Outputf(Info, format, v...)
}

func (logger *DefaultLogger) Warn(v ...interface{}) {
	logger.Output(Warn, v...)
}

func (logger *DefaultLogger) Warnf(format string, v ...interface{}) {
	logger.Outputf(Warn, format, v...)
}

func (logger *DefaultLogger) Error(v ...interface{}) {
	logger.Output(Error, v...)
}

func (logger *DefaultLogger) Errorf(format string, v ...interface{}) {
	logger.Outputf(Error, format, v...)
}

func (logger *DefaultLogger) Fatal(v ...interface{}) {
	logger.Output(Fatal, v...)
}

func (logger *DefaultLogger) Fatalf(format string, v ...interface{}) {
	logger.Outputf(Fatal, format, v...)
}
