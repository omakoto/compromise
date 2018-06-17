package adapters

import (
	"github.com/omakoto/compromise/src/compromise/compmisc"
	"github.com/omakoto/go-common/src/common"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func getUniqueName(command string) string {
	return toShellSafeName(common.MustGetBinName()) + "_" + toShellSafeName(command)
}

func specFileName(command string) string {
	dir := compmisc.SpecPath
	err := os.MkdirAll(dir, 0700)
	common.Checkf(err, "unable to create directory %s", dir)
	return filepath.Join(dir, "compspec_"+getUniqueName(command)+".txt")
}

func SaveSpec(command, spec string) string {
	file := specFileName(command)
	err := ioutil.WriteFile(file, []byte(spec), 0600)
	common.Checkf(err, "unable to write to %s", file)
	return file
}

func getFuncName(command string) string {
	return "__compromise_" + getUniqueName(command) + "_completion"
}

func escapeCommandName(commandName string, realEscaper func(string) string) string {
	if strings.HasPrefix(commandName, `@"`) && strings.HasSuffix(commandName, `"`) {
		return commandName[2 : len(commandName)-1]
	}

	return realEscaper(commandName)
}
