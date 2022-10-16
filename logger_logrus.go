package msgbroker_nsq

import (
	"github.com/sirupsen/logrus"
)

type loggerLogrus struct {
}

func (l *loggerLogrus) Output(calldepth int, s string) error {
	if len(s) < 3 {
		return nil
	}

	level := s[0:3]
	switch level {
	case "INF":
		logrus.Info(s)
	case "ERR":
		logrus.Error(s)
	case "WRN":
		logrus.Warn(s)
	case "DBG":
		logrus.Debug(s)
	default:
		logrus.Info(s)
	}

	return nil
}
