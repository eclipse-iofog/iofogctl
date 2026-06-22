package auth

import (
	"crypto/rand"
	"math/big"

	inputvalidate "github.com/eclipse-iofog/iofogctl/internal/validate"
)

const tempPasswordLength = 16

const (
	tempPasswordUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	tempPasswordLower   = "abcdefghijklmnopqrstuvwxyz"
	tempPasswordDigits  = "0123456789"
	tempPasswordSpecial = "!@#$%^&*"
)

var generateTempPasswordFn = generateTempPassword

func generateTempPassword() (string, error) {
	all := tempPasswordUpper + tempPasswordLower + tempPasswordDigits + tempPasswordSpecial

	required, err := randomChars([]string{
		tempPasswordUpper,
		tempPasswordDigits,
		tempPasswordSpecial,
	})
	if err != nil {
		return "", err
	}

	password := make([]byte, 0, tempPasswordLength)
	password = append(password, required...)
	for len(password) < tempPasswordLength {
		ch, err := randomCharFrom(all)
		if err != nil {
			return "", err
		}
		password = append(password, ch)
	}

	if err := shuffleBytes(password); err != nil {
		return "", err
	}

	result := string(password)
	if err := inputvalidate.ValidatePasswordComplexity(result); err != nil {
		return "", err
	}
	return result, nil
}

func randomChars(charsets []string) ([]byte, error) {
	out := make([]byte, 0, len(charsets))
	for _, charset := range charsets {
		ch, err := randomCharFrom(charset)
		if err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	return out, nil
}

func randomCharFrom(charset string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
	if err != nil {
		return 0, err
	}
	return charset[n.Int64()], nil
}

func shuffleBytes(buf []byte) error {
	for i := len(buf) - 1; i > 0; i-- {
		j, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		buf[i], buf[j.Int64()] = buf[j.Int64()], buf[i]
	}
	return nil
}
