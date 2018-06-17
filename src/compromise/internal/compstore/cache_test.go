package compstore

import (
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compenv"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCashe(t *testing.T) {
	source := []compromise.Candidate{
		compromise.NewCandidateBuilder().Build(),
		compromise.NewCandidateBuilder().Value("x").Build(),
		compromise.NewCandidateBuilder().Help("help!").Build(),
		compromise.NewCandidateBuilder().Raw(true).Build(),
		compromise.NewCandidateBuilder().Hidden(true).Build(),
		compromise.NewCandidateBuilder().Force(true).Build(),
		compromise.NewCandidateBuilder().Continues(true).Build(),
		compromise.NewCandidateBuilder().NeedsHelp(true).Build(),

		compromise.NewCandidateBuilder().Value("v").Help("h").Raw(true).Hidden(true).Force(true).Continues(true).NeedsHelp(true).Build(),
	}

	compenv.CacheFilename = "/tmp/compromise-test-cache.dat"

	err := CacheCandidates(source)
	if err != nil {
		assert.FailNow(t, "Failed to save to cache: %s", err)
	}

	loaded, err := LoadCandidates()
	if err != nil {
		assert.FailNow(t, "Failed to load from cache: %s", err)
	}

	assert.Equal(t, source, loaded)
}
