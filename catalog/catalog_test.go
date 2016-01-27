package catalog

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitCatalog(t *testing.T) {
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "..")
	catalog := InitCatalog(path)
	if catalog == nil {
		t.Error("Did not read catalog")
	}
}
