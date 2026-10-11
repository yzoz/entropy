// Package crypto provides Ed25519 signing, address derivation and key
// (de)serialization primitives.
package crypto

import "encoding/base64"

func b64(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

func unb64(s string) ([]byte, error) { return base64.StdEncoding.DecodeString(s) }
