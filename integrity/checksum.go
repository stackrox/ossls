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
