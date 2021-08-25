// Copyright (C) 2019-2021, Xiongfa Li.
// @author xiongfa.li
// @version V1.0
// Description:

package nevetrace

import "github.com/xfali/xlog"

type JaegerLogger struct {
	logger xlog.Logger
}

func NewLogger() *JaegerLogger {
	return &JaegerLogger{
		logger: xlog.GetLogger().WithDepth(2),
	}
}

// Error logs a message at error priority
func (l *JaegerLogger) Error(msg string) {
	l.logger.Errorln(msg)
}

// Infof logs a message at info priority
func (l *JaegerLogger) Infof(msg string, args ...interface{}) {
	l.logger.Infof(msg, args...)
}

func (l *JaegerLogger) Debugf(msg string, args ...interface{}) {
	l.logger.Debugf(msg, args...)
}
