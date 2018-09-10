package audit

import (
	"io/ioutil"

	"github.com/joshdk/licensor"
	"github.com/joshdk/licensor/spdx"
)

func extractLicenseFromFileBody(filename string) ([]spdx.License, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	match := licensor.Best(body)

	return []spdx.License{match.License}, nil
}
