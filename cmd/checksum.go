package cmd

import (
	"fmt"

	"github.com/stackrox/ossls/integrity"
)

func Checksum(filename string) (string, error) {
	return integrity.Checksum(filename)
}

func ChecksumPrint(filename string, checksum string) {
	fmt.Printf("%s %s\n", filename, checksum)
}
