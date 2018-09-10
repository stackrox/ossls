package audit

import (
	"encoding/json"
	"io/ioutil"
)

// f1 represents the structure of a package.json file, where the license data
// is specified under the "license" property, as a single string.
//
// {
//   "license": "BSD"
// }
type f1 struct {
	License string `json:"license"`
}

func (f *f1) Licenses() []info {
	return []info{
		{
			Type: f.License,
		},
	}
}

// f2 represents the structure of a package.json file, where the license data
// is specified under the "license" property, as a list of strings.
//
// {
//   "license": [
//     "BSD",
//     "MIT"
//   ]
// }
type f2 struct {
	License []string `json:"license"`
}

func (f *f2) Licenses() []info {
	infos := make([]info, len(f.License))
	for index, license := range f.License {
		infos[index] = info{
			Type: license,
		}
	}
	return infos
}

// f3 represents the structure of a package.json file, where the license data
// is specified under the "license" property, as a single object with "type"
// and "url" properties.
//
// {
//   "license": {
//     "type": "BSD",
//     "url":  "https://github.com/.../blob/master/LICENSE"
//   }
// }
type f3 struct {
	License info `json:"license"`
}

func (f *f3) Licenses() []info {
	return []info{
		f.License,
	}
}

// f4 represents the structure of a package.json file, where the license data
// is specified under the "licenses" property, as a list of objects with "type"
// and "url" properties.
//
// {
//   "license": [
//     {
//       "type": "BSD",
//       "url":  "https://github.com/.../blob/master/LICENSE"
//     },
//     {
//       "type": "MIT",
//       "url":  "https://github.com/.../blob/master/LICENSE"
//     }
//   ]
// }
type f4 struct {
	Ls []info `json:"licenses"`
}

func (f *f4) Licenses() []info {
	return f.Ls
}

type format interface {
	Licenses() []info
}

func extractLicenseFromPackageJson(filename string) ([]info, error) {
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Create one of each format type, and attempt to unmarshal into each one.
	// Return licenses from the first successful format.
	formats := []format{&f1{}, &f2{}, &f3{}, &f4{}}
	for _, format := range formats {
		if err = json.Unmarshal(body, &format); err == nil {
			return format.Licenses(), nil
		}
	}

	// Return the last error from json.Unmarshal
	return nil, err
}
