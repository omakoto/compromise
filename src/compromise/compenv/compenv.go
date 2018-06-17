package compenv

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
	// Whether to suppress verbose output or not.
	Quiet = getBoolEnv("COMPROMISE_QUIET", false)

	// Whether to enable debug log or not.
	DebugEnabled = getBoolEnv("COMPROMISE_DEBUG", false)

	// Whether to enable timing log or not.
	Time = getBoolEnv("COMPROMISE_TIME", false)

	// Log filename.
	LogFile = utils.FirstNonEmpty(os.Getenv("COMPROMISE_LOG_FILE"), "/tmp/compromise.log")

	// Whether to use color
	UseColor = !getBoolEnv("COMPROMISE_NO_COLOR", false)

	HelpStartEscape = utils.FirstNonEmpty(os.Getenv("COMPROMISE_HELP_START"), "\x1b[36m")
	HelpEndEscape   = utils.FirstNonEmpty(os.Getenv("COMPROMISE_HELP_END"), "\x1b[0m")

	// Whether to do case insensitive match or not.
	IgnoreCase = getBoolEnv("COMPROMISE_IGNORE_CASE", true)

	// Treat underscores and hyphens interchangeably.
	MapCase = getBoolEnv("COMPROMISE_MAP_CASE", true)

	// Bell style, not used yet.
	BellStyle = os.Getenv("COMPROMISE_BELL_STYPE")

	// (Bash only) show this many candidates initially. Double tab TAB to see more candidates.
	FirstMaxCandidates = utils.ParseInt(os.Getenv("COMPROMISE_FIRST_MAX_CANDIDATES"), 10, 40)

	// At most show this many candidates.
	MaxCandidates = utils.ParseInt(os.Getenv("COMPROMISE_MAX_CANDIDATES"), 10, 2000)

	// Home is a home directory path.
	Home = os.Getenv("HOME")

	// Persistent storage filename.
	DoublePressTimeout = time.Duration(utils.ParseInt(os.Getenv("COMPROMISE_DOUBLE_PRESS_TIMEOUT_MS"), 10, 500)) * time.Millisecond

	// On bash, show help for at most this many candidates.
	BashHelpMaxCandidates = utils.ParseInt(os.Getenv("COMPROMISE_BASH_HELP_MAX"), 10, 20)

	// On bash, do not execute bind commands
	BashSkipBind = getBoolEnv("COMPROMISE_BASH_SKIP_BINDS", false)

	// Compromise dot directory path.
	CompDir = utils.FirstNonEmpty(os.Getenv("COMPROMISE_DIR"), filepath.Join(Home, ".compromise"))

	// Persistent storage filename.
	StoreFilename = path.Join(CompDir, "lastcommand.json")

	// Cached candidates file. If the last completion was very recent and the command line is the same,
	// compromise reuses cached candidates.
	// Set "" to disable cache.
	CacheFilename = path.Join(CompDir, "lastcandidates.dat")

	// Timeout for the cache.
	CacheTimeout = time.Duration(utils.ParseInt(os.Getenv("COMPROMISE_CACHE_TIMEOUT_MS"), 10, 1000)) * time.Millisecond

	// Where spec files are stored.
	SpecPath = path.Join(CompDir, "spec")

	// Whether to use fzf or not
	UseFzf = getBoolEnv("COMPROMISE_USE_FZF", false)

	// Filename of the FZF executable.
	FzfBinName = utils.FirstNonEmpty(os.Getenv("COMPROMISE_FZF_BIN"), "fzf")

	// Whether to show candidates in reverse order on FZF.
	FzfFlip = getBoolEnv("COMPROMISE_FZF_FLIP", false)

	// Extra options to pass to fzf.
	FzfOptions = utils.FirstNonEmpty(os.Getenv("COMPROMISE_FZF_OPTIONS"), "")
)
