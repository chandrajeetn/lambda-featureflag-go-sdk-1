package logger

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

var loggerInst = &logrus.Entry{}

type LogFormat struct {
	TimestampFormat string
}

// WriterHook is a hook that writes logs of specified LogLevels to specified Writer
type WriterHook struct {
	Writer       io.Writer
	LogLevels    []logrus.Level
	LogFormatter logrus.Formatter
}

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (hook *WriterHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

// Levels define on which log levels this hook would trigger
func (hook *WriterHook) Levels() []logrus.Level {
	return hook.LogLevels
}

func (f *LogFormat) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer

	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	b.WriteString("time=\"")
	b.WriteString(entry.Time.UTC().Format(time.RFC3339) + "\"")

	b.WriteString(" ")

	b.WriteString("level=")
	fmt.Fprint(b, "\""+strings.ToUpper(entry.Level.String())+"\"")

	b.WriteString(" ")

	b.WriteString("message=")
	if entry.Message != "" {
		b.WriteString(fmt.Sprintf("\"%s\"", entry.Message))
	}

	//b.WriteString("producerType=\"service\" ")
	//b.WriteString("producerName=\"hps\" ")
	//b.WriteString("hostname=\"" + os.Getenv("HOSTNAME") + "\" ")

	//mdc
	for key, value := range entry.Data {
		b.WriteString(" " + key)
		b.WriteString("=")
		b.WriteString(fmt.Sprintf("\"%v\"", value))

	}
	//end mdc

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func init() {

	logrusInstance := logrus.New()
	// override logType to line when when explicitly specied
	if os.Getenv("logType") == "JSON" {
		logrusInstance.Formatter = &logrus.JSONFormatter{}
	} else {
		logrusInstance.Formatter = new(LogFormat)
	}

	logrusInstance.Level = getLogLevel()
	enableFileLogging, _ := strconv.ParseBool(os.Getenv("ENABLE_FILE_LOGGING"))
	if enableFileLogging {
		logrusInstance.SetOutput(ioutil.Discard)
		filePath := "/opt/logs/lhps/info.log"
		var lumberjackLogger, lumberjackLoggerErr *lumberjack.Logger

		lumberjackLogger = &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    100,
			MaxBackups: 10,
			MaxAge:     7,
			LocalTime:  true,
			Compress:   true,
		}

		errorFilePath := "/opt/logs/lhps/error.log"
		lumberjackLoggerErr = &lumberjack.Logger{
			Filename:   errorFilePath,
			MaxSize:    100,
			MaxBackups: 10,
			MaxAge:     7,
			LocalTime:  true,
			Compress:   true,
		}
		// Add a hook with all logging levels for Stdout logging
		mwErr := io.MultiWriter(os.Stderr, lumberjackLoggerErr)
		logrusInstance.AddHook(&WriterHook{ // Send logs with level higher than warning to stderr
			Writer: mwErr,
			LogLevels: []logrus.Level{
				logrus.PanicLevel,
				logrus.FatalLevel,
				logrus.ErrorLevel,
				logrus.WarnLevel,
			},
		})

		mwOut := io.MultiWriter(os.Stdout, lumberjackLogger)

		logrusInstance.AddHook(&WriterHook{
			Writer: mwOut,
			LogLevels: []logrus.Level{
				logrus.PanicLevel,
				logrus.FatalLevel,
				logrus.ErrorLevel,
				logrus.WarnLevel,
				logrus.InfoLevel,
				logrus.DebugLevel,
			},
		})
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGHUP)
		go func() {
			for {
				time.Sleep(10 * time.Second)
				<-c
				lumberjackLogger.Rotate()
				lumberjackLoggerErr.Rotate()
			}
		}()
	} else {
		logrusInstance.SetOutput(os.Stdout)
	}
	loggerInst = logrusInstance.WithFields(logrus.Fields{"hostName": os.Getenv("HOSTNAME")})
}

func GetLogger() (logger *logrus.Entry) {
	return loggerInst.WithFields(logrus.Fields{})
}

func getLogLevel() logrus.Level {
	envLogLevel := os.Getenv("LOG_LEVEL")
	if envLogLevel == "" {
		// if none is defined, then default it is info level
		envLogLevel = "info"
	}
	switch strings.ToLower(envLogLevel) {
	case "info":
		return logrus.InfoLevel
	case "debug":
		return logrus.DebugLevel
	case "panic":
		return logrus.PanicLevel
	case "error":
		return logrus.ErrorLevel
	case "fatal":
		return logrus.FatalLevel
	case "warn":
		return logrus.WarnLevel
	case "trace":
		return logrus.TraceLevel
	default:
		return logrus.InfoLevel
	}
}
