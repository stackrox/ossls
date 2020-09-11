package resolver

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

func FindLicenseFiles(dirname string, matcher Matcher) ([]string, error) {
	var foundFiles []string

	entries, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, errors.Wrapf(err, "reading directory %s", dirname)
	}
	for _, entry := range entries {
		if entry.Mode()&os.ModeType != 0 {
			continue
		}
		if matcher(entry.Name()) {
			foundFiles = append(foundFiles, filepath.Join(dirname, entry.Name()))
		}
	}
	return foundFiles, nil
}
