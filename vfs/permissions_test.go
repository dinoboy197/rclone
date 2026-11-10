package vfs

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rclone/rclone/backend/memory"
	"github.com/rclone/rclone/fs"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fs/object"
	"github.com/rclone/rclone/vfs/vfscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// putPermissionObject puts an object with the given name into the memory fs.
func putPermissionObject(ctx context.Context, t *testing.T, f fs.Fs, name string) fs.Object {
	t.Helper()
	contents := "hello, permissions"
	obji := object.NewStaticObjectInfo(name, time.Now(), int64(len(contents)), true, nil, nil)
	obj, err := f.Put(ctx, strings.NewReader(contents), obji)
	require.NoError(t, err)
	return obj
}

// setPermMeta sets metadata on an object via the optional SetMetadataer
// interface.
func setPermMeta(ctx context.Context, t *testing.T, obj fs.Object, meta fs.Metadata) {
	t.Helper()
	do, ok := obj.(fs.SetMetadataer)
	require.True(t, ok, "object should support SetMetadata")
	require.NoError(t, do.SetMetadata(ctx, meta))
}

// asFile stats a VFS node at path and returns it as a *File.
func asFile(t *testing.T, vfs *VFS, name string) *File {
	t.Helper()
	node, err := vfs.Stat(name)
	require.NoError(t, err)
	file, ok := node.(*File)
	require.True(t, ok)
	return file
}

// TestVFSPermissionsInMetadata checks that when --vfs-permissions is enabled
// the file mode, uid and gid are read from and written to object metadata.
func TestVFSPermissionsInMetadata(t *testing.T) {
	ctx, _ := fs.AddConfig(context.Background())
	f, err := memory.NewFs(ctx, "memory", ":memory:", configmap.Simple{})
	require.NoError(t, err)

	obj := putPermissionObject(ctx, t, f, "permfile")
	setPermMeta(ctx, t, obj, fs.Metadata{
		"mode":   "100755",
		"uid":    "500",
		"gid":    "501",
		"potato": "existing",
	})

	opt := &vfscommon.Options{
		Permissions: true,
		FilePerms:   0o600,
		DirPerms:    0o700,
		UID:         1000,
		GID:         1001,
	}
	vfs := New(ctx, f, opt)
	t.Cleanup(func() { vfs.Shutdown() })

	file := asFile(t, vfs, "permfile")
	assert.Equal(t, os.FileMode(0o755), file.Mode())
	assert.Equal(t, uint32(500), file.Uid())
	assert.Equal(t, uint32(501), file.Gid())

	// chmod writes the new mode to the object metadata
	require.NoError(t, vfs.Chmod("permfile", 0o600))
	file = asFile(t, vfs, "permfile")
	assert.Equal(t, os.FileMode(0o600), file.Mode())

	// chmod must preserve pre-existing user metadata
	meta, err := fs.GetMetadata(ctx, file.o)
	require.NoError(t, err)
	assert.Equal(t, "existing", meta["potato"])
	assert.Equal(t, "100600", meta["mode"])

	// chown writes the new uid/gid to the object metadata
	require.NoError(t, vfs.Chown("permfile", 777, 888))
	file = asFile(t, vfs, "permfile")
	assert.Equal(t, uint32(777), file.Uid())
	assert.Equal(t, uint32(888), file.Gid())
	meta, err = fs.GetMetadata(ctx, file.o)
	require.NoError(t, err)
	assert.Equal(t, "777", meta["uid"])
	assert.Equal(t, "888", meta["gid"])
	assert.Equal(t, "existing", meta["potato"])
}

// TestVFSPermissionsBackwardCompat checks that without --vfs-permissions the
// metadata is ignored and the default permissions are used, even when the
// object has permission metadata.
func TestVFSPermissionsBackwardCompat(t *testing.T) {
	ctx, _ := fs.AddConfig(context.Background())
	f, err := memory.NewFs(ctx, "memory", ":memory:", configmap.Simple{})
	require.NoError(t, err)

	obj := putPermissionObject(ctx, t, f, "permfile_bc")
	setPermMeta(ctx, t, obj, fs.Metadata{"mode": "100755", "uid": "500", "gid": "501"})

	opt := &vfscommon.Options{
		Permissions: false,
		FilePerms:   0o600,
		UID:         1000,
		GID:         1001,
	}
	vfs := New(ctx, f, opt)
	t.Cleanup(func() { vfs.Shutdown() })

	file := asFile(t, vfs, "permfile_bc")
	assert.Equal(t, os.FileMode(0o600), file.Mode())
	assert.Equal(t, uint32(1000), file.Uid())
	assert.Equal(t, uint32(1001), file.Gid())

	// with permissions disabled, chmod is unsupported
	assert.Equal(t, ENOSYS, vfs.Chmod("permfile_bc", 0o600))
}
