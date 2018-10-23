package cmd

import (
	"fmt"
)

func ChecksumPrint(filename string, field string, checksum string) {
	switch field {
	case "":
		fmt.Printf("%s %s\n", filename, checksum)
	default:
		fmt.Printf("%s [%s] %s\n", filename, field, checksum)
	}
}
