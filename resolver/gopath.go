package resolver

import (
	"os/exec"
	"sync"
)

var (
	goPath     string
	goPathInit sync.Once
)

// GoPath retrieves the value of the GOPATH variable
func GoPath() string {
	goPathInit.Do(func() {
		path, err := getGoPath()
		if err != nil {
			panic(err)
		}
		goPath = path
	})
	return goPath
}

func getGoPath() (string, error) {
	output, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
