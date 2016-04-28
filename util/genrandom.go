package util

import (
	"io"
	mrand "math/rand"
	"time"
	"crypto/rand"
	"fmt"
)

var stdCharsType1 = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*")
var stdCharsType2 = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")

//generate new random password
func NewRandomPassword(length int) string {
	return rand_char(length, stdCharsType1)
}

//generate new app key for new clients
func NewAppKey() string {
	return rand_char(16, stdCharsType2)
}

//generate new refresh token uuid
func GenerateNewRefreshToken() (string, error) {
	uuid := make([]byte, 16)
	n, err := io.ReadFull(rand.Reader, uuid)
	if n != len(uuid) || err != nil {
		return "", err
	}
	// variant bits; see section 4.1.1
	uuid[8] = uuid[8] &^ 0xc0 | 0x80
	// version 4 (pseudo-random); see section 4.1.3
	uuid[6] = uuid[6] &^ 0xf0 | 0x40
	return fmt.Sprintf("%x", uuid), nil
}

//generate random number in range
func GenerateRandomNumber(min, max int) int {
	mrand.Seed(time.Now().UTC().UnixNano())
	return mrand.Intn(max - min) + min
}

//random char string with length and bytes sets
func rand_char(length int, chars []byte) string {
	new_random := make([]byte, length)
	// storage for random bytes
	random_data := make([]byte, length + (length / 4))
	clen := byte(len(chars))
	maxrb := byte(256 - (256 % len(chars)))
	i := 0
	for {
		if _, err := io.ReadFull(rand.Reader, random_data); err != nil {
			panic(err)
		}
		for _, c := range random_data {
			if c >= maxrb {
				continue
			}
			new_random[i] = chars[c % clen]
			i++
			if i == length {
				return string(new_random)
			}
		}
	}
	panic("unreachable")
}
