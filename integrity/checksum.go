package integrity

import (
	"crypto/sha256"
	"encoding/json"
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

func ChecksumField(data interface{}) (string, error) {
	serialized, err := json.MarshalIndent(data, "", "  ")
	serialized = append(serialized, '\n')
	if err != nil {
		return "", err
	}

	return ChecksumBytes(serialized), nil
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

func VerifyField(data interface{}, checksum string) (bool, string, error) {
	actual, err := ChecksumField(data)
	if err != nil {
		return false, "", err
	}
	return checksum == actual, actual, nil
}

func VerifyBytes(data []byte, checksum string) (bool, string) {
	actual := ChecksumBytes(data)
	return checksum == actual, actual
}
