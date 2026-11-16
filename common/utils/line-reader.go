package utils

import (
	"bufio"
	"bytes"
	"os"
	"strings"

	"github.com/pkg/errors"
)

// utf8BOM 是 UTF-8 字节序标记，用命名常量代替散落的字节字面量。
var (
	utf8BOM       = []byte{0xef, 0xbb, 0xbf}
	utf8BOMString = string(utf8BOM)
)

func RemoveBOM(raw []byte) []byte {
	if len(raw) > len(utf8BOM) && bytes.HasPrefix(raw, utf8BOM) {
		return raw[len(utf8BOM):]
	}
	return raw
}

func RemoveBOMForString(raw string) string {
	if len(raw) > len(utf8BOMString) && strings.HasPrefix(raw, utf8BOMString) {
		return raw[len(utf8BOMString):]
	}
	return raw
}

func FileLineReader(file string) (chan []byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, errors.Errorf("failed to read file: %s", err)
	}

	reader := bufio.NewReader(f)
	outC := make(chan []byte)
	go func() {
		defer f.Close()
		defer close(outC)
		bomHandled := NewBool(false)
		for {
			lineRaw, err := BufioReadLine(reader)
			if err != nil {
				break
			}
			raw := bytes.TrimSpace(lineRaw)
			if !bomHandled.IsSet() {
				raw = RemoveBOM(raw)
				bomHandled.Set()
			}
			outC <- raw
		}
	}()

	return outC, nil
}
