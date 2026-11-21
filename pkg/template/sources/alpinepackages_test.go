package sources

import (
	"testing"

	"github.com/csmith/latest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAlpinePackages(t *testing.T) {
	packages := map[string]*latest.AlpinePackageInfo{
		"curl":    {Name: "curl", Version: "8.10.1-r0", Dependencies: []string{"libssl3"}},
		"libssl3": {Name: "libssl3", Version: "3.3.2-r0"},
	}

	got, err := resolveAlpinePackages(packages, "curl")
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"curl": "8.10.1-r0", "libssl3": "3.3.2-r0"}, got)

	_, err = resolveAlpinePackages(packages, "missing")
	assert.Error(t, err)
}
