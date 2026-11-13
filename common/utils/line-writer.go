package utils

import (
	"io/fs"
	"os"
)

type fileLineWriter struct {
	*os.File
	flag      int
	firstLine bool
}

// readLastByte 用独立的只读句柄读取 path 中偏移 offset 处的字节。
func readLastByte(path string, offset int64) (byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	buf := make([]byte, 1)
	if _, err := f.ReadAt(buf, offset); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func NewFileLineWriter(file string, flag int, perm fs.FileMode) (*fileLineWriter, error) {
	f, err := os.OpenFile(file, flag, perm)
	if err != nil {
		return nil, err
	}
	// 追加时，假如最后一个字符不是换行符，则添加换行符
	if flag&os.O_APPEND != 0 {
		state, err := f.Stat()
		if err != nil {
			return nil, err
		}
		// 空文件没有“最后一个字符”，直接追加即可；否则会在偏移 -1 处
		// ReadAt 失败并误写一个前导换行符。另外 f 可能是只写句柄，
		// 无法 ReadAt，需要用单独的只读句柄读取最后一个字节。
		if state.Size() > 0 {
			last, err := readLastByte(file, state.Size()-1)
			if err != nil {
				// 无法读取时退回“补一个换行符”，避免破坏既有的可追加文件。
				last = 0
			}
			if last != '\n' {
				_, err := f.WriteString("\n")
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return &fileLineWriter{File: f, flag: flag, firstLine: true}, nil
}

func (w *fileLineWriter) WriteLine(line []byte) (n int, err error) {
	if w.firstLine {
		w.firstLine = false
	} else {
		n, err = w.File.WriteString("\n")
	}
	n, err = w.File.Write(line)
	if err != nil {
		return
	}
	return
}

func (w *fileLineWriter) WriteLineString(line string) (n int, err error) {
	if w.firstLine {
		w.firstLine = false
	} else {
		n, err = w.File.WriteString("\n")
	}
	n, err = w.File.WriteString(line)
	if err != nil {
		return
	}
	return
}
