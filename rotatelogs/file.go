package rotatelogs

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var (
	defaultFormater           = new(FileFormat)
	defaultMaxSize      int64 = 100 * 1024 * 1024
	defaultFormatLayout       = "20060102150405"
)

type FileFormater interface {
	Format(time.Time) string
	IsFormat(string) bool
}

type FileFormat struct {
	Prefix string
}

func (f *FileFormat) Format(t time.Time) string {
	return f.Prefix + t.Format(defaultFormatLayout) + ".log"
}

func (f *FileFormat) IsFormat(s string) bool {
	if !strings.HasPrefix(s, f.Prefix) {
		return false
	}
	_, err := time.Parse(defaultFormatLayout, strings.TrimSuffix(s, ".log")[len(f.Prefix):])
	return err == nil
}

func New(dir string, formater FileFormater, maxSize int64) (io.Writer, error) {
	var err error
	file := &file{
		format:  defaultFormater,
		maxSize: defaultMaxSize,
	}

	if formater != nil {
		file.format = formater
	}

	if maxSize > 0 {
		file.maxSize = maxSize
	}

	if dir == "" {
		file.absDir, err = filepath.Abs("log")
		return file, err
	}

	file.absDir, err = filepath.Abs(dir)
	return file, err
}

type file struct {
	absDir  string
	format  FileFormater
	file    *os.File
	size    int64
	maxSize int64
}

// Write 实现 io.Writer 接口
func (f *file) Write(data []byte) (n int, err error) {
	// 文件存在, 写入
	if f.file != nil {
		return f.write(data)
	}

	// 文件不存在, 检查是否有历史日志文件, 没有会新建
	err = f.checkFile()
	if err != nil {
		panic(err)
	}

	return f.write(data)
}

func (f *file) write(data []byte) (n int, err error) {
	// 检查数据写入后文件大小是否会超过阈值
	if f.size+int64(len(data)) <= f.maxSize {
		f.size += int64(len(data))
		return f.file.Write(data)
	}

	// 如果会超过阈值, 新建文件
	file, err := f.newFile()
	if err != nil {
		panic(err)
	}
	f.file.Close()
	f.file = file

	// 写入
	f.size = int64(len(data))
	return f.file.Write(data)
}

// checkFile 检查是否有历史日志文件, 没有会新建
func (f *file) checkFile() (err error) {
	// 文件对象已存在, 略过检查
	if f.file != nil {
		return nil
	}

	// 保证文件夹存在
	err = os.MkdirAll(f.absDir, os.ModePerm)
	if err != nil {
		return errors.New("创建日志文件夹失败")
	}

	// 打开日志文件夹
	files, err := ioutil.ReadDir(filepath.Join(f.absDir))
	if err != nil {
		return err
	}
	// 按创建时间排序, 最后创建(最新)的文件下标最小(为0)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	// 遍历所有文件, 查看是否有旧的日志文件
	for i := 0; i < len(files); i++ {
		file := files[i]
		// 前缀检查
		if !f.format.IsFormat(file.Name()) {
			continue
		}

		// 检查文件大小
		if file.Size() >= f.maxSize {
			continue
		}

		// 找到符合的文件, 打开返回
		fileName := filepath.Join(f.absDir, file.Name())
		f.file, err = openFile(fileName)
		if err != nil {
			return err
		}
		f.size = file.Size()
		return nil
	}

	// 没有找到符合要求的附件, 新建
	file, err := f.newFile()
	if err != nil {
		return err
	}

	f.file = file
	return nil
}

// newFile 以只追加的形式新建文件
func (f *file) newFile() (*os.File, error) {
	fileName := filepath.Join(f.absDir, f.format.Format(time.Now()))
	return openFile(fileName)
}

// openFile 以只追加的形式打开文件
func openFile(fileName string) (*os.File, error) {
	return os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
}
