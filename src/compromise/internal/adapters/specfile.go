package adapters

import (
	"github.com/omakoto/compromise-go/src/compromise/internal/compmisc"
	"github.com/omakoto/go-common/src/common"
	"io/ioutil"
	"os"
	"path/filepath"
)

func specFileName(command string) string {
	dir := compmisc.SpecPath
	err := os.MkdirAll(dir, 0700)
	common.Checkf(err, "unable to create directory %s", dir)
	return filepath.Join(dir, "compspec_" + toShellSafeName(command) + ".txt")
}

func saveSpec(command, spec string) string {
	file := specFileName(command)
	err := ioutil.WriteFile(file, []byte(spec), 0600)
	common.Checkf(err, "unable to write to %s", file)
	return file
}
