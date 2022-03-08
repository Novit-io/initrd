package config

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

var (
	encoding = base64.RawStdEncoding
)

func PasswordHashFromSeed(seed, pass []byte) string {
	h := pbkdf2.Key(pass, seed, 2048, 32, sha512.New)
	return encoding.EncodeToString(h)
}

func PasswordHash(pass []byte) (hashedPassWithSeed string) {
	seed := make([]byte, 10) // 8 bytes min by the RFC recommendation
	_, err := rand.Read(seed)
	if err != nil {
		panic(err) // we do not expect this to fail...
	}
	return JoinSeedAndHash(seed, PasswordHashFromSeed(seed, pass))
}

func JoinSeedAndHash(seed []byte, hash string) string {
	return encoding.EncodeToString(seed) + ":" + hash
}

func CheckPassword(hashedPassWithSeed string, pass []byte) (ok bool) {
	parts := strings.SplitN(hashedPassWithSeed, ":", 2)

	encodedSeed := parts[0]
	encodedHash := parts[1]

	seed, err := encoding.DecodeString(encodedSeed)
	if err != nil {
		return false
	}

	return encodedHash == PasswordHashFromSeed(seed, pass)
}
