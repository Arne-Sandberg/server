package hashers

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/riesinger/freecloud/utils"
	"golang.org/x/crypto/scrypt"
)

const (
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
		return "", err
	}

	hashs := base64.StdEncoding.EncodeToString(hashb)
	salts := base64.StdEncoding.EncodeToString(saltb)

	return fmt.Sprintf("$%s$%d$%d$%d$%s%s", ScryptHashID, recommendedN, recommendedp, recommendedr, salts, hashs), nil

}

func ParseScryptStub(password string) (salt, hash []byte, N, r, p int, err error) {
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
		return
	}

	N = int(n64)

	r64, err = strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		return
	}

	r = int(r64)

	p64, err = strconv.ParseInt(parts[2], 10, 0)
	if err != nil {
		return
	}

	p = int(p64)

	salt, err = base64.StdEncoding.DecodeString(parts[3])
	if err != nil {
		return
	}

	hash, err = base64.StdEncoding.DecodeString(parts[4])
	return
}
