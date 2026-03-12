package progress

import (
	"os"
)

// ProgressWriter 带进度跟踪的写入器
type ProgressWriter struct {
	File     *os.File
	TaskID   string
	FromByte int64
	ToByte   int64
}

func (pw *ProgressWriter) Write(p []byte) (n int, err error) {
	n, err = pw.File.WriteAt(p, pw.FromByte)
	if err != nil {
		return n, err
	}
	pw.FromByte += int64(n)

	return n, nil
}
