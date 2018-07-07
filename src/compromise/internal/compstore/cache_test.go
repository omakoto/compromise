package compstore

import (
	"github.com/omakoto/compromise/src/compromise"
	"github.com/omakoto/compromise/src/compromise/compenv"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCashe(t *testing.T) {
	source := []compromise.Candidate{
		compromise.NewCandidate(),
		compromise.NewCandidate().SetValue("x"),
		compromise.NewCandidate().SetHelp("help!"),
		compromise.NewCandidate().SetRaw(true),
		compromise.NewCandidate().SetHidden(true),
		compromise.NewCandidate().SetForce(true),
		compromise.NewCandidate().SetContinues(true),

		compromise.NewCandidate().SetValue("v").SetHelp("h").SetRaw(true).SetHidden(true).SetForce(true).SetContinues(true),
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
