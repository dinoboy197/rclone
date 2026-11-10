package fs

import (
	"fmt"
	"os"
	"strconv"
)

// Metadata key names used to store POSIX permissions in object metadata.
//
// These match the keys used by the local backend (see backend/local) so that
// permissions round-trip between remotes with --metadata.
const (
	MetadataKeyMode  = "mode"
	MetadataKeyUID   = "uid"
	MetadataKeyGID   = "gid"
	MetadataKeyRDev  = "rdev"
	MetadataKeyATime = "atime"
	MetadataKeyMTime = "mtime"
	MetadataKeyBTime = "btime"
)

// Permissions is the POSIX permission data stored in an object's metadata and
// interpreted by the VFS (rclone mount).
type Permissions struct {
	Mode os.FileMode // file mode in Unix st_mode format, incl. type bits
	UID  int         // user ID of owner
	GID  int         // group ID of owner
}

// PermissionsFromMetadata reads POSIX permissions from a metadata map.
//
// It parses the "mode" (octal), "uid" and "gid" (decimal) keys, in the same
// formats as the local backend. It returns found=false when none of these keys
// are present, and an error when a present key can't be parsed.
func PermissionsFromMetadata(m Metadata) (perms Permissions, found bool, err error) {
	if s, ok := m[MetadataKeyMode]; ok {
		v, err := strconv.ParseUint(s, 8, 32)
		if err != nil {
			return perms, true, fmt.Errorf("failed to get permissions from metadata: failed to parse metadata %q = %q: %w", MetadataKeyMode, s, err)
		}
		perms.Mode = os.FileMode(v)
		found = true
	}
	if s, ok := m[MetadataKeyUID]; ok {
		v, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return perms, true, fmt.Errorf("failed to get permissions from metadata: failed to parse metadata %q = %q: %w", MetadataKeyUID, s, err)
		}
		perms.UID = int(v)
		found = true
	}
	if s, ok := m[MetadataKeyGID]; ok {
		v, err := strconv.ParseUint(s, 10, 32)
		if err != nil {
			return perms, true, fmt.Errorf("failed to get permissions from metadata: failed to parse metadata %q = %q: %w", MetadataKeyGID, s, err)
		}
		perms.GID = int(v)
		found = true
	}
	return perms, found, nil
}

// ToMetadata returns the permissions as a metadata map using the same keys
// and formats as the local backend: "mode" in octal as the raw st_mode, "uid"
// and "gid" in decimal.
func (p Permissions) ToMetadata() Metadata {
	return Metadata{
		MetadataKeyMode: fmt.Sprintf("%0o", uint32(p.Mode)),
		MetadataKeyUID:  strconv.FormatInt(int64(p.UID), 10),
		MetadataKeyGID:  strconv.FormatInt(int64(p.GID), 10),
	}
}
