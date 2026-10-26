package exec

import (
	"fmt"
	"testing"
)

func TestIsNormalExecClose(t *testing.T) {
	err1005 := fmt.Errorf("%s: ", execCloseCodeString(1005))
	err1000 := fmt.Errorf("%s: done", execCloseCodeString(1000))

	if !isNormalExecClose(err1000, false) {
		t.Fatal("expected code 1000 to be normal")
	}
	if isNormalExecClose(err1005, false) {
		t.Fatal("expected code 1005 without user close to be abnormal")
	}
	if !isNormalExecClose(err1005, true) {
		t.Fatal("expected code 1005 with user close to be normal")
	}
}
