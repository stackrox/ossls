package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackrox/ossls/integrity"
)

func ChecksumCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "checksum",
		Short: "Calculate checksum for a file",
		RunE: func(_ *cobra.Command, args []string) error {
			var (
				filename string
				field    string
				checksum string
				err      error
			)

			switch len(args) {
			case 1:
				filename = args[0]
				checksum, err = integrity.Checksum(filename)
			case 2:
				filename, field = args[0], args[1]
				checksum, err = integrity.ChecksumFileField(filename, field)
			default:
				return errors.New("bad argument count")
			}
			if err != nil {
				return err
			}

			ChecksumPrint(filename, field, checksum)
			return nil
		},
	}
}

func ChecksumPrint(filename string, field string, checksum string) {
	switch field {
	case "":
		fmt.Printf("%s %s\n", filename, checksum)
	default:
		fmt.Printf("%s [%s] %s\n", filename, field, checksum)
	}
}
