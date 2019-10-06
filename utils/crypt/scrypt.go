package crypt

import (
	"TraefikAccessControl/utils"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"golang.org/x/crypto/scrypt"
)

const (
	saltLength = 16

	recommendedN = 16384
	recommendedr = 8
	recommendedp = 1

	scryptHashLength = 32

	ScryptHashID = "s1"
)

var ErrInvalidScryptStub = errors.New("auth: invalid scrypt stub")

func HashScrypt(plaintext string) (hash string, err error) {
	passwordb := []byte(plaintext)
	saltb := []byte(utils.RandomString(saltLength))

	hashb, err := scrypt.Key(passwordb, saltb, recommendedN, recommendedr, recommendedp, scryptHashLength)
	if err != nil {
		return "", errors.Wrap(err, "could not hash password")
	}

	hashs := base64.StdEncoding.EncodeToString(hashb)
	salts := base64.StdEncoding.EncodeToString(saltb)

	return fmt.Sprintf("$%s$%d$%d$%d$%s$%s", ScryptHashID, recommendedN, recommendedr, recommendedp, salts, hashs), nil
}

func ValidateScryptPassword(plaintext, hashed string) (valid bool, err error) {
	// First, parse the stub of the hash to get the scrypt parameters
	salt, oldHash, N, r, p, err := parseScryptStub(hashed)
	if err != nil {
		err = errors.Wrap(err, "could not get the password stub")
		return
	}

	hash, err := scrypt.Key([]byte(plaintext), salt, N, r, p, scryptHashLength)
	if err != nil {
		err = errors.Wrap(err, "could not hash password")
		return
	}
	// Check if the old hash is the same as the new one
	if base64.StdEncoding.EncodeToString(hash) == base64.StdEncoding.EncodeToString(oldHash) {
		valid = true
		err = nil
		return
	}

	return
}

func parseScryptStub(password string) (salt, hash []byte, N, r, p int, err error) {
	// First, do some cheap sanity checking
	if len(password) < 10 || !strings.HasPrefix(password, fmt.Sprintf("$%s$", ScryptHashID)) {
		err = ErrInvalidScryptStub
		return
	}

	// strip the $<ScryptHashID>$, then split into parts
	parts := strings.Split(password[4:], "$")
	// We need N, r, p, salt and the hash
	if len(parts) < 5 {
		err = ErrInvalidScryptStub
		return
	}

	var n64, r64, p64 int64

	n64, err = strconv.ParseInt(parts[0], 10, 0)
	if err != nil {
		err = errors.Wrap(err, "could not parse scrypt parameters")
		return
	}

	N = int(n64)

	r64, err = strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		err = errors.Wrap(err, "could not parse scrypt parameters")
		return
	}

	r = int(r64)

	p64, err = strconv.ParseInt(parts[2], 10, 0)
	if err != nil {
		err = errors.Wrap(err, "could not parse scrypt parameters")
		return
	}

	p = int(p64)

	salt, err = base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		err = errors.Wrap(err, "could not parse scrypt salt")
		return
	}

	hash, err = base64.StdEncoding.DecodeString(parts[4])
	if err != nil {
		err = errors.Wrap(err, "could not parse scrypt hash")
	}
	return
}
