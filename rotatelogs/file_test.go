package rotatelogs

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSlice(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), time.Now().Format("20060102150405"))
	defer func() {
		os.RemoveAll(tmpDir)
	}()
	w, err := New(tmpDir, nil, 10)
	assert.Equal(t, nil, err)

	logrus.SetOutput(w)
	for i := 0; i < 10; i++ {
		w.Write([]byte("write data"))
		time.Sleep(1 * time.Second)
	}
	files, err := ioutil.ReadDir(tmpDir)
	assert.Equal(t, nil, err)
	assert.Equal(t, 10, len(files))
	for _, file := range files {
		assert.EqualValues(t, 10, file.Size())
	}
}
