package hook

import (
	"io"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

var (
	defaultWriter    = os.Stderr
	defaultFormatter = &logrus.TextFormatter{DisableColors: true}
)

// New 创建并返回一个 hook
//
// 该过程会检查 Writer 和 Formatter 是否有效, 无效则使用默认选项:
// Writer 默认输出到 os.Stderr;
// Formatter 默认为	&logrus.TextFormatter{DisableColors: true}
func New(w io.Writer, format logrus.Formatter, levels ...logrus.Level) *hook {
	h := &hook{
		mu: &sync.Mutex{},
	}
	h.SetWriter(w)
	h.SetLevels(levels)
	h.SetFormatter(format)
	return h
}

// hook 实现了 logrus 中的 hook 接口, 用于向不同的 writer 输出不同的格式数据
type hook struct {
	writer    io.Writer
	levels    []logrus.Level
	formatter logrus.Formatter
	mu        *sync.Mutex
}

// Fire 实现 logrus 中的 hook 接口, 用于输出
func (h *hook) Fire(entry *logrus.Entry) error {
	msg, err := h.formatter.Format(entry)
	if err != nil {
		return err
	}
	h.mu.Lock()
	_, err = h.writer.Write(msg)
	h.mu.Unlock()
	return err
}

// Levels 实现 logrus 中的 hook 接口, 返回当前 hook 支持的日志等级
func (h *hook) Levels() []logrus.Level {
	return h.levels
}

// SetWriter 设置默认的输出位置
func (h *hook) SetWriter(w io.Writer) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 设置默认的输出位置
	if w == nil {
		h.writer = defaultWriter
		return
	}

	h.writer = w
}

// SetFormatter 设置输出格式
func (h *hook) SetFormatter(formatter logrus.Formatter) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 设置默认的输出格式
	if formatter == nil {
		h.formatter = defaultFormatter
		return
	}

	h.formatter = formatter
}

// SetLevels 设置日志等级
func (h *hook) SetLevels(levels []logrus.Level) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// 检查并设置日志等级
	for _, level := range levels {
		if level.String() == "unknown" {
			continue
		}
		h.levels = append(h.levels, level)
	}

	if len(h.levels) == 0 {
		h.levels = append(h.levels, logrus.InfoLevel)
	}
}
