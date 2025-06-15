package toolset

import (
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"fmt"
	"strings"
)

func ConvertPubDestToB32(pubDestBase64 string) (string, error) {
	safeBase64 := strings.NewReplacer("~", "+", "-", "/").Replace(pubDestBase64)

	// Pad the base64 string if necessary
	if mod := len(safeBase64) % 4; mod != 0 {
		safeBase64 += strings.Repeat("=", 4-mod)
	}

	pubDestBytes, err := base64.StdEncoding.DecodeString(safeBase64)
	if err != nil {
		return "", fmt.Errorf("base64 decode error: %w", err)
	}

	hash := sha256.Sum256(pubDestBytes)

	b32 := base32.StdEncoding.EncodeToString(hash[:])
	b32 = strings.ToLower(strings.TrimRight(b32, "="))

	return b32 + ".b32.i2p", nil
}
