package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/yaklang/yaklang/common/log"
	"testing"
)

func TestLastLine(t *testing.T) {
	trueCase := map[string]string{
		`aasdfas
asdf
asdf
asdf
aaaa`: "aaaa",
		`aaaa`: "aaaa",
	}

	for k, v := range trueCase {
		if v != string(LastLine([]byte(k))) {
			t.FailNow()
		}
		log.Infof("%s 's last line is %s", k, v)
	}
}

func TestParseStringToVisible(t *testing.T) {
	for i := range make([]int, 256) {
		res := ParseStringToVisible(string([]byte{byte(i)}))
		fmt.Printf("%v (0x%02x)\n", res, i)
	}
}

func TestPrettifyListFromStringSplit(t *testing.T) {
	assert.Equal(t, []string{"a", "b", "c", "d", "e", "f", "g"}, PrettifyListFromStringSplitEx("a,b,c|d|e,f|g", ",", "|"))
	assert.Equal(t, []string{"a", "b", "c", "d", "e", "f", "g"}, PrettifyListFromStringSplitEx("a,b,c d|e,f|g", ",", "|", " "))
	assert.Equal(t, []string{"a", "b", "c"}, PrettifyListFromStringSplitEx("abc", ""))
	assert.Equal(t, []string{"a", "b", "c"}, PrettifyListFromStringSplitEx("a b c", " "))
}

func TestPrettifyJoin(t *testing.T) {
	assert.Equal(t, "a,b,c,d,e,f,g", PrettifyJoin(",", "a", "b", "c", "d", "e", "f", "g"))
	assert.Equal(t, "a,b,c,f,g", PrettifyJoin(",", "a", "b", "c", "", "", "f", "g"))
	assert.Equal(t, "a,b,c,f", PrettifyJoin(",", "a", "b", "c", "", "", "f", ""))
	assert.Equal(t, "b,c,f,g", PrettifyJoin(",", "", "b", "c", "", "", "f", "g"))
	assert.Equal(t, "c,f,g", PrettifyJoin(",", "", "", "c", "", "", "f", "g"))
	assert.Equal(t, "c,f", PrettifyJoin(",", "", "", "c", "", "", "f", ""))
	assert.Equal(t, "c,f,c,f,c,f,c,f,c,f,c,f", PrettifyJoin(",", "", "", "c", "", "", "f", "", "c", "f", "c", "f", "c", "f", "c", "f", "c", "f"))
	assert.Equal(t, "c,f", PrettifyShrinkJoin(",", "", "", "c", "", "", "f", "", "c", "f", "c", "f", "c", "f", "c", "f", "c", "f"))
}

func TestExtractTitleFromHTMLTitle(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{"plain", "<title>abc</title>", "abc"},
		{"nested in document", "<html><head><title>Hello</title></head></html>", "Hello"},
		{"whitespace preserved", "<title>  x  </title>", "  x  "},
		{"empty returns default", "<title></title>", "<default>"},
		{"missing returns default", "<html><body>no title</body></html>", "<default>"},
		// 非法 UTF-8 字节会被转义成 \xNN，必须完整保留而不是按原长度截断。
		{"invalid utf8 byte", "<title>a\xffb</title>", `a\xffb`},
		// 控制字符同理。
		{"control byte", "<title>a\x01b</title>", `a\x01b`},
		// 闭合标签可以是 <title>（无斜杠），不能吞掉末尾字符。
		{"closing title without slash", "<title>abc<title>", "abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ExtractTitleFromHTMLTitle(tt.html, "<default>"))
		})
	}
}
