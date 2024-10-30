package huabanv3

import "testing"

func TestDownloadBoard(t *testing.T) {
	err := DownloadBoard(94004345, 2, false)
	if err != nil {
		t.Fatalf("%v", err)
	}
}
