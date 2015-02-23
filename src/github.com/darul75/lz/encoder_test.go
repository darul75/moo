package lz

import (
	"fmt"
	"testing"
)

func TestEncoding(t *testing.T) {
	original := "hello1hello2hello3hello4hello5hello6hello7hello8hello9helloAhelloBhelloChelloDhelloEhelloF"
	data := &Data{original, ""}

	str := data.Encode()
	fmt.Println(str)
	end := data.Decode()

	if end != original {
		t.Errorf("OMG")
	}
}
