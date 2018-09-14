package integrity

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
)

func Checksum(filename string) (string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return ChecksumBytes(data), nil
}

func ChecksumBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

func Verify(filename string, checksum string) (bool, string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return false, "", err
	}

	ok, actual := VerifyBytes(data, checksum)
	return ok, actual, nil
}

func VerifyBytes(data []byte, checksum string) (bool, string) {
	actual := ChecksumBytes(data)
	return checksum == actual, actual
}
