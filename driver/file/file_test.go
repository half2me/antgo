package file

import (
	"testing"
	"os"
	"time"
	"fmt"
)

func TestGetAntCaptureFile(t *testing.T) {
	f_t := GetAntCaptureFile("testFile.cap")
	if err := f_t.Create(); err != nil {
		t.Error("Could not open test capture file!")
	}

	defer os.Remove("testFile.cap")
	defer f_t.Close()

	f_t.Write([]byte("ANT+"))
	f_t.Close()

	f_t.Open()
	if a, e := f_t.Read(); e != nil {
		t.Error(e.Error())
	} else {
		if string(a.Data) != "ANT+" {t.Error("Data is corrupt")}
		if a.Timestamp.Sub(time.Now()) > time.Second {t.Error("Timestamp is corrupt")}
	}
}
