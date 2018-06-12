package compmisc

import (
	"github.com/omakoto/go-common/src/utils"
	"os"
	"path"
	"path/filepath"
	"time"
)

func getBoolEnv(name string, def bool) bool {
	d := 0
	if def {
		d = 1
	}
	return utils.ParseInt(os.Getenv(name), 10, d) == 1
}

var (
	// Whether to enable debug log or not.
	DebugEnabled = getBoolEnv("COMPROMISE_DEBUG", false)

	// Whether to enable verbose output or not.
	Verbose = getBoolEnv("COMPROMISE_VERBOSE", true)

	// Whether to enable timing log or not.
	Time = getBoolEnv("COMPROMISE_TIME", false)

	// Log filename.
	LogFile = utils.FirstNonEmpty(os.Getenv("COMPROMISE_LOG_FILE"), "/tmp/compromise.log")

	// Whether to use color
	UseColor = !getBoolEnv("COMPROMISE_NO_COLOR", false)

	// Whether to do case insensitive match or not.
	IgnoreCase = getBoolEnv("COMPROMISE_IGNORE_CASE", true)

	// Treat underscores and hyphens interchangeably.
	MapCase = getBoolEnv("COMPROMISE_MAP_CASE", true)

	// Bell style, not used yet.
	BellStyle = os.Getenv("COMPROMISE_BELL_STYPE")

	// At most show this many candidates.
	MaxCandidates = utils.ParseInt(os.Getenv("COMPROMISE_MAX_CANDIDATES"), 10, 1000)

	// Home is a home directory path.
	Home = os.Getenv("HOME")

	// Persistent storage filename.
	DoublePressTimeout = time.Duration(utils.ParseInt(os.Getenv("COMPROMISE_DOUBLE_PRESS_TIMEOUT_MS"), 10, 300)) * time.Millisecond

	// On bash, show help for at most this many candidates.
	BashHelpMaxCandidates = utils.ParseInt(os.Getenv("COMPROMISE_BASH_HELP_MAX"), 10, 20)

	// On bash, do not execute bind commands
	BashSkipBind = getBoolEnv("COMPROMISE_BASH_SKIP_BINDS", false)

	// Compromise dot directory path.
	CompDir = utils.FirstNonEmpty(os.Getenv("COMPROMISE_DIR"), filepath.Join(Home, ".compromise"))

	// Persistent storage filename.
	StoreFilename = path.Join(CompDir, "lastcommand.json")

	// Where spec files are stored.
	SpecPath = path.Join(CompDir, "spec")
)
