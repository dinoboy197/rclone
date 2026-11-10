//go:build linux

package mount

import (
	"context"
	"strings"
	"testing"
	"time"

	"bazil.org/fuse"
	"github.com/rclone/rclone/backend/memory"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fs/object"
	vfs "github.com/rclone/rclone/vfs"
	"github.com/rclone/rclone/vfs/vfscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetattrChmodChown proves that the FUSE Setattr handler routes mode and
// owner changes through to the VFS Chmod/Chown, which store them in the object
// metadata when --vfs-permissions is enabled.
func TestSetattrChmodChown(t *testing.T) {
	ctx, _ := fs.AddConfig(context.Background())
	f, err := memory.NewFs(ctx, "memory", ":memory:", configmap.Simple{})
	require.NoError(t, err)

	contents := "hello, setattr"
	obji := object.NewStaticObjectInfo("permfile", time.Now(), int64(len(contents)), true, nil, nil)
	_, err = f.Put(ctx, strings.NewReader(contents), obji)
	require.NoError(t, err)

	opt := &vfscommon.Options{
		Permissions: true,
		FilePerms:   0o600,
		DirPerms:    0o700,
		UID:         1000,
		GID:         1001,
	}
	theVFS := vfs.New(ctx, f, opt)
	t.Cleanup(func() { theVFS.Shutdown() })

	node, err := theVFS.Stat("permfile")
	require.NoError(t, err)
	vfile, ok := node.(*vfs.File)
	require.True(t, ok)
	mfile := &File{File: vfile}

	// chmod 644 through the FUSE Setattr layer
	err = mfile.Setattr(ctx, &fuse.SetattrRequest{
		Valid: fuse.SetattrMode,
		Mode:  0o644,
	}, &fuse.SetattrResponse{})
	require.NoError(t, err)
	meta, err := fs.GetMetadata(ctx, vfile.DirEntry())
	require.NoError(t, err)
	assert.Equalf(t, "100644", meta[fs.MetadataKeyMode], "chmod did not reach the object metadata, Setattr probably ignored the mode")

	// chown 1234:5678 through the FUSE Setattr layer
	err = mfile.Setattr(ctx, &fuse.SetattrRequest{
		Valid: fuse.SetattrUid | fuse.SetattrGid,
		Uid:   1234,
		Gid:   5678,
	}, &fuse.SetattrResponse{})
	require.NoError(t, err)
	meta, err = fs.GetMetadata(ctx, vfile.DirEntry())
	require.NoError(t, err)
	assert.Equal(t, "1234", meta[fs.MetadataKeyUID])
	assert.Equal(t, "5678", meta[fs.MetadataKeyGID])
}
