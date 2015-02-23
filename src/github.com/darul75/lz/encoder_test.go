package lz

import (
	"testing"
)

func TestEncoding(t *testing.T) {
	original := "test"
	data := &Data{original, ""}

	data.Encode()
	end := data.Decode()

	if end != original {
		t.Errorf("OMG")
	}
}
