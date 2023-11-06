package utility

import (
	"encoding/hex"
	"hash/crc32"
	"schemastash/types"
)

func VersionHash(schematic types.Schematic, version types.Version) string {
	hash := crc32.NewIEEE()
	hash.Write([]byte(schematic.ID))
	hash.Write([]byte(schematic.CreatedAt))
	hash.Write([]byte(version.CreatedAt))
	hash.Write([]byte(version.Data))
	return hex.EncodeToString(hash.Sum(nil))
}
