package mmclogd

import (
	"os"
	"testing"
)

const ()

func TestRmFile(t *testing.T) {
	fn := "/tmp/tempfile"
	f, err := os.Create(fn)
	if err != nil {
		t.Errorf("Could not create file, error: %v", err)
		return
	}
	f.Close()
	if _, err = os.Stat(fn); os.IsNotExist(err) {
		t.Errorf("File did not get created, error: %v", err)
		return
	}
	if err = rmFile(fn); err != nil {
		t.Errorf("Error during remove file: %v", err)
		return
	}
	if _, err = os.Stat(fn); !os.IsNotExist(err) {
		t.Errorf("File did not get removed, error: %v", err)
		return
	}
}
