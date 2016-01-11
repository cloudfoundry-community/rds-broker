package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitSecrets(t *testing.T) {
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "..")
	secrets := InitSecrets(path)
	if secrets == nil {
		t.Error("Did not read catalog")
	}
}
