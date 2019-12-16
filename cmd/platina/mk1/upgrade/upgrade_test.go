// +build +brokenTests

package upgrade

import (
	"os"
	"testing"
)

const (
	TFTPserver = "192.168.101.127" //Invader7 for now
)

func TestGetFile(t *testing.T) {
	fn := "LIST"
	s := DfltSrv
	v := DfltVer
	tftp := false
	n, err := getFile(s, v, tftp, fn)
	if err != nil {
		t.Errorf("HTTP: Error downloading: %v", err)
		return
	}
	if n < 10 {
		t.Errorf("HTTP: Error file too small: %v", err)
		return
	}
	if _, err = os.Stat(fn); os.IsNotExist(err) {
		t.Errorf("HTTP: File did not get created, error: %v", err)
		return
	}
	if err = rmFile(fn); err != nil {
		t.Errorf("HTTP: File did not get removed, error: %v", err)
		return
	}

	fn = "LIST"
	s = TFTPserver
	v = DfltVer
	tftp = true
	n, err = getFile(s, v, tftp, fn)
	if err != nil {
		t.Errorf("TFTP: Error downloading: %v", err)
		return
	}
	if n < 10 {
		t.Errorf("TFTP: Error file too small: %v", err)
		return
	}
	if _, err = os.Stat(fn); os.IsNotExist(err) {
		t.Errorf("TFTP: File did not get created, error: %v", err)
		return
	}
	if err = rmFile(fn); err != nil {
		t.Errorf("TFTP: File did not get removed, error: %v", err)
		return
	}
}

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
