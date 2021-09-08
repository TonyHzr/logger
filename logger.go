package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/TonyHzr/logger/hook"
	"github.com/TonyHzr/logger/rotatelogs"
	"github.com/sirupsen/logrus"
)

type Option struct {
	WriteToStd     bool
	WriteToFile    bool
	LogDir         string
	MaxFileSize    int64
	OutputLevel    string
	StdOutFormat   logrus.Formatter
	FileOutFormat  logrus.Formatter
	FileNameFormat rotatelogs.FileFormater
}

var (
	timeFormat = "0102 15:04:05000"
	stdOutput  = &logrus.TextFormatter{
		FullTimestamp:    true,
		ForceColors:      true,
		TimestampFormat:  timeFormat,
		CallerPrettyfier: callerPrettyfier,
	}
	fileOutput = &logrus.JSONFormatter{
		TimestampFormat:  timeFormat,
		CallerPrettyfier: callerPrettyfier,
	}
)

func callerPrettyfier(frame *runtime.Frame) (function string, file string) {
	file = fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
	return
}

func SetLevel(outLevel string) error {
	level, err := logrus.ParseLevel(outLevel)
	if err != nil {
		return err
	}

	logrus.SetLevel(level)
	return nil
}

func Init(option Option) error {
	err := SetLevel(option.OutputLevel)
	if err != nil {
		return err
	}

	logrus.SetReportCaller(true)

	// 标准输出
	if option.WriteToStd {
		logrus.SetOutput(os.Stderr)

		// 输出格式
		if option.StdOutFormat != nil {
			logrus.SetFormatter(option.StdOutFormat)
		} else {
			logrus.SetFormatter(stdOutput)
		}
	} else {
		logrus.SetOutput(io.Discard)
	}

	if !option.WriteToFile {
		return nil
	}

	// 文件输出
	for _, level := range logrus.AllLevels {
		writer, err := rotatelogs.New(option.LogDir, &rotatelogs.FileFormat{Prefix: level.String() + "_"}, option.MaxFileSize)
		if err != nil {
			return err
		}

		hook := hook.New(writer, fileOutput, level)
		logrus.AddHook(hook)
	}

	return nil
}
