package fs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionsFromMetadata(t *testing.T) {
	for _, test := range []struct {
		name     string
		metadata Metadata
		perms    Permissions
		found    bool
		err      bool
	}{
		{
			name: "empty",
		},
		{
			name:     "full",
			metadata: Metadata{"mode": "100664", "uid": "500", "gid": "501"},
			perms:    Permissions{Mode: os.FileMode(0o100664), UID: 500, GID: 501},
			found:    true,
		},
		{
			name:     "dir mode",
			metadata: Metadata{"mode": "40755"},
			perms:    Permissions{Mode: os.FileMode(0o40755)},
			found:    true,
		},
		{
			name:     "uid only",
			metadata: Metadata{"uid": "1000"},
			perms:    Permissions{UID: 1000},
			found:    true,
		},
		{
			name:     "bad mode",
			metadata: Metadata{"mode": "zzz"},
			err:      true,
			found:    true,
		},
		{
			name:     "bad uid",
			metadata: Metadata{"uid": "abc"},
			err:      true,
			found:    true,
		},
		{
			name:     "bad gid",
			metadata: Metadata{"gid": "9x"},
			err:      true,
			found:    true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			perms, found, err := PermissionsFromMetadata(test.metadata)
			if test.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, test.found, found)
			assert.Equal(t, test.perms, perms)
		})
	}
}

func TestPermissionsToMetadata(t *testing.T) {
	p := Permissions{Mode: os.FileMode(0o100644), UID: 500, GID: 501}
	got := p.ToMetadata()
	assert.Equal(t, "100644", got["mode"])
	assert.Equal(t, "500", got["uid"])
	assert.Equal(t, "501", got["gid"])
}

func TestPermissionsRoundTrip(t *testing.T) {
	p := Permissions{Mode: os.FileMode(0o100755), UID: 1000, GID: 1001}
	got, found, err := PermissionsFromMetadata(p.ToMetadata())
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, p, got)
}
