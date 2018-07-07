package comptest

import (
	"github.com/omakoto/compromise/src/compromise/compdebug"
	"github.com/omakoto/go-common/src/common"
	"os"
	"os/exec"
	"regexp"
)

type injectedOutput struct {
	pattern *regexp.Regexp
	output  string
}

var injectedOutputs []injectedOutput

// Execute a command and return the stdout.
func ExecAndGetStdout(command string) ([]byte, error) {
	compdebug.Debugf("Executing: %q\n", command)

	for _, i := range injectedOutputs {
		if i.pattern.MatchString(command) {
			return []byte(i.output), nil
		}
	}
	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()

	if err != nil {
		common.Warnf("Command execution error: command=%q error=%s", command, err)
	}
	return output, err
}

// Inject an output for a command for testing.
func InjectCommandOutput(pattern, output string) {
	injectedOutputs = append(injectedOutputs, injectedOutput{regexp.MustCompile(pattern), output})
}
