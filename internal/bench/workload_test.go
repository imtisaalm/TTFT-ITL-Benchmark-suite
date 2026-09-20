package bench

import (
	"strings"
	"testing"
)

func TestUniquePrefixDefault(t *testing.T) {
	p := Prompt("base", 7, false)
	if !strings.HasPrefix(p, "[benchmark-request=7]") {
		t.Fatalf("request marker should lead prompt")
	}
}

func TestSharedPrefixMode(t *testing.T) {
	p := Prompt("base", 7, true)
	if !strings.HasPrefix(p, "base") || !strings.HasSuffix(p, "[benchmark-request=7]") {
		t.Fatalf("shared prefix mode should preserve base prompt prefix")
	}
}
