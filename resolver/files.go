package resolver

import (
	"os"
	"path/filepath"
)

func FindLicenseFiles(dirname string, patterns []string) []string {
	var (
		foundFiles = []string{}
		//patterns   = []string{
		//	"*AUTHOR*",
		//	"*COPYING*",
		//	"*LICENSE*",
		//	"*LICENCE*",
		//	"*NOTICE*",
		//	"*author*",
		//	"*copying*",
		//	"*license*",
		//	"*License*",
		//	"*notice*",
		//	"package.json",
		//}
	)

	for _, pattern := range patterns {
		glob := filepath.Join(dirname, pattern)
		matches, _ := filepath.Glob(glob)
		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			foundFiles = append(foundFiles, match)
		}
	}

	return foundFiles
}
