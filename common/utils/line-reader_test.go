package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveBOM(t *testing.T) {
	bom := []byte{0xef, 0xbb, 0xbf}
	tests := []struct {
		name string
		in   []byte
		want []byte
	}{
		{"empty", nil, nil},
		{"no bom", []byte("hello"), []byte("hello")},
		{"bom with content", append(append([]byte{}, bom...), "hello"...), []byte("hello")},
		{"bom in the middle is untouched", []byte("a\xef\xbb\xbfb"), []byte("a\xef\xbb\xbfb")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, RemoveBOM(tt.in))
			assert.Equal(t, string(tt.want), RemoveBOMForString(string(tt.in)))
		})
	}
}
