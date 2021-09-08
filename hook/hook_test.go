package hook

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func getBufAndFormat() (*bytes.Buffer, *logrus.TextFormatter) {
	buf := bytes.NewBuffer(nil)
	format := &logrus.TextFormatter{
		DisableTimestamp: true,
		DisableColors:    true,
	}
	return buf, format
}

// TestLevels 测试不同等级的日志输出
func TestLevels(t *testing.T) {
	buf, format := getBufAndFormat()

	defer func() {
		// 捕获 PanicLevel 日志
		if err := recover(); err != nil {
			assert.Equal(t, fmt.Sprintf("level=%s msg=\"send to buf\"\n", logrus.PanicLevel.String()), buf.String())
		}
	}()

	for _, level := range []logrus.Level{
		logrus.TraceLevel,
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.FatalLevel,
		logrus.PanicLevel,
	} {
		// 新建日志, 设置输出等级, 抛弃输出
		log := logrus.New()
		log.SetLevel(level)
		log.SetOutput(ioutil.Discard)

		// 添加 hook
		hook := New(buf, format, level)
		log.AddHook(hook)
		buf.Reset()
		log.Log(level, "send to buf")
		assert.Equal(t, fmt.Sprintf("level=%s msg=\"send to buf\"\n", level.String()), buf.String())
	}
}

// TestMutilLevels 多种等级的日志写入同一个 Writer
func TestMutilLevels(t *testing.T) {
	buf, format := getBufAndFormat()

	// 新建日志, 设置输出等级, 抛弃输出
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(ioutil.Discard)

	// 添加 hook
	hook := New(buf, format, logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel)
	log.AddHook(hook)
	buf.Reset()

	log.Log(logrus.DebugLevel, "send to buf")
	log.Log(logrus.InfoLevel, "send to buf")
	log.Log(logrus.WarnLevel, "send to buf")

	msgs := strings.Split(buf.String(), "\n")
	assert.Equal(t, "level=debug msg=\"send to buf\"", msgs[0])
	assert.Equal(t, "level=info msg=\"send to buf\"", msgs[1])
	assert.Equal(t, "level=warning msg=\"send to buf\"", msgs[2])
	assert.Equal(t, "", msgs[3])
}

// TestMutilPrint 一条日志写入多个 Writer
func TestMutilWriters(t *testing.T) {
	buf, format := getBufAndFormat()

	// 新建日志, 设置输出等级, 抛弃输出
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetOutput(ioutil.Discard)

	// 添加 hook
	hook1 := New(buf, format, logrus.InfoLevel)
	log.AddHook(hook1)
	buf.Reset()
	hook2 := New(buf, format, logrus.InfoLevel)
	log.AddHook(hook2)
	buf.Reset()
	hook3 := New(buf, format, logrus.InfoLevel)
	log.AddHook(hook3)
	buf.Reset()

	log.Log(logrus.InfoLevel, "send to buf")

	msgs := strings.Split(buf.String(), "\n")
	assert.Equal(t, "level=info msg=\"send to buf\"", msgs[0])
	assert.Equal(t, "level=info msg=\"send to buf\"", msgs[1])
	assert.Equal(t, "level=info msg=\"send to buf\"", msgs[2])
	assert.Equal(t, "", msgs[3])
}

// TestParrelle 测试多线程输出
func TestParrelle(t *testing.T) {
	buf, format := getBufAndFormat()
	N := 100

	// 新建日志, 设置输出等级, 抛弃输出
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetOutput(ioutil.Discard)

	// 添加 hook
	hook := New(buf, format, logrus.InfoLevel)
	log.AddHook(hook)

	sw := &sync.WaitGroup{}
	for i := 0; i < runtime.NumCPU(); i++ {
		sw.Add(1)
		go func(cpuNumber int) {
			for cnt := 0; cnt < N; cnt++ {
				log.Infof("%d send %d", cpuNumber, cnt)
			}
			sw.Done()
		}(i)
	}
	sw.Wait()

	msgCnt := map[int]int{}
	re := regexp.MustCompile(`(?m)^level=info msg="(\d+) send (\d+)"$`)
	msgs := re.FindAllStringSubmatch(buf.String(), -1)
	for _, msg := range msgs {
		cpuNumber, err := strconv.Atoi(msg[1])
		assert.Equal(t, nil, err)

		cnt, err := strconv.Atoi(msg[2])
		assert.Equal(t, nil, err)

		assert.Equal(t, msgCnt[cpuNumber], cnt)
		msgCnt[cpuNumber]++
	}

	assert.Equal(t, 8, len(msgCnt))
	for _, v := range msgCnt {
		assert.Equal(t, 100, v)
	}

	rst := strings.Split(buf.String(), "\n")
	assert.Equal(t, runtime.NumCPU()*N+1, len(rst))
	assert.Equal(t, "", rst[len(rst)-1])
	for i := 0; i < len(rst)-1; i++ {

	}
}
