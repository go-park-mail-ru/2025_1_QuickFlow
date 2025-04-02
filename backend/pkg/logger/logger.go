package logger

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func init() {
	Log = logrus.New()
	Log.SetFormatter(&CustomFormatter{})
	Log.SetOutput(os.Stdout)
	Log.SetLevel(logrus.InfoLevel)
}

func Info(args ...interface{}) {
	logWithContext().Info(args...)
}

func Warn(args ...interface{}) {
	logWithContext().Warn(args...)
}

func Error(args ...interface{}) {
	logWithContext().Error(args...)
}

func Debug(args ...interface{}) {
	logWithContext().Debug(args...)
}

func logWithContext() *logrus.Entry {
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		return Log.WithFields(logrus.Fields{
			"package":  "unknown",
			"function": "anonymous",
		})
	}

	funcObj := runtime.FuncForPC(pc)
	if funcObj == nil {
		return Log.WithFields(logrus.Fields{
			"package":  "unknown",
			"function": "anonymous",
		})
	}

	funcPath := funcObj.Name()
	lastSlash := strings.LastIndex(funcPath, "/")
	if lastSlash < 0 {
		lastSlash = 0
	}

	parts := strings.Split(funcPath[lastSlash:], ".")
	var packageName, funcName string

	switch len(parts) {
	case 0:
		packageName = "unknown"
		funcName = "anonymous"
	case 1:
		packageName = parts[0]
		funcName = "anonymous"
	default:
		packageName = parts[0]
		funcName = strings.Join(parts[1:], ".")
	}

	funcName = strings.TrimLeft(funcName, "*")

	return Log.WithFields(logrus.Fields{
		"package":  packageName,
		"function": funcName,
	})
}

type CustomFormatter struct{}

func (f *CustomFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor string
	switch entry.Level {

	case logrus.InfoLevel:
		levelColor = "\033[34m" // Синий

	case logrus.WarnLevel:
		levelColor = "\033[33m" // Жёлтый

	case logrus.ErrorLevel:
		levelColor = "\033[31m" // Красный

	case logrus.FatalLevel, logrus.PanicLevel:
		levelColor = "\033[35m" // Фиолетовый

	default:
		levelColor = "\033[0m" // Без цвета
	}

	timestamp := time.Now().Format("2006-01-02T15:04:05")
	packageName := fmt.Sprintf("\033[33m[%s]\033[0m", entry.Data["package"])
	funcName := fmt.Sprintf("\033[36m[%s]\033[0m", entry.Data["function"])
	level := fmt.Sprintf("%s[%s]\033[0m", levelColor, strings.ToUpper(entry.Level.String()))

	logMessage := fmt.Sprintf("%s[%s]%s%s %s\n",
		level,
		timestamp,
		packageName,
		funcName,
		entry.Message,
	)

	return []byte(logMessage), nil
}
