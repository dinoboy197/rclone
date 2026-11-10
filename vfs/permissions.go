package vfs

import (
	"context"
	"os"

	"github.com/rclone/rclone/fs"
)

// permissionInfo holds POSIX permissions resolved from an object's metadata.
//
// The has* flags record which of mode, uid or gid were present in the
// metadata, so that missing fields fall back to the VFS defaults.
type permissionInfo struct {
	mode    os.FileMode // permission and special bits (low 12) of the stored mode
	hasMode bool
	uid     uint32
	hasUID  bool
	gid     uint32
	hasGID  bool
}

// permissionsFromEntry reads the POSIX permissions from the metadata of the
// entry.
//
// It returns has=false when the entry doesn't support metadata or none of the
// permission keys are present, and an error when a present key can't be
// parsed.
func permissionsFromEntry(ctx context.Context, entry fs.DirEntry) (pi permissionInfo, has bool, err error) {
	metadata, err := fs.GetMetadata(ctx, entry)
	if err != nil {
		return pi, false, err
	}
	if metadata == nil {
		return pi, false, nil
	}
	p, _, err := fs.PermissionsFromMetadata(metadata)
	if err != nil {
		return pi, false, err
	}
	if _, ok := metadata[fs.MetadataKeyMode]; ok {
		pi.mode = p.Mode & 0o7777
		pi.hasMode = true
	}
	if _, ok := metadata[fs.MetadataKeyUID]; ok {
		pi.uid = uint32(p.UID)
		pi.hasUID = true
	}
	if _, ok := metadata[fs.MetadataKeyGID]; ok {
		pi.gid = uint32(p.GID)
		pi.hasGID = true
	}
	return pi, pi.hasMode || pi.hasUID || pi.hasGID, nil
}
