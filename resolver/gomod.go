package resolver

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
)

type GoModProject struct {
	Path    string
	Version string
	Dir     string
	Replace *GoModProject
}

func ProjectsFromGoModFile(filename string) ([]GoModProject, error) {
	cmd := exec.Command("go", "list", "-json", "-m", "all")
	cmd.Dir = filepath.Dir(filename)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create pipe for redirecting stdout")
	}
	defer func() {
		_ = stdout.Close()
	}()

	var projects []GoModProject
	errC := make(chan error, 1)
	go func() {
		jsonDec := json.NewDecoder(stdout)
		var project GoModProject
		err := jsonDec.Decode(&project)
		for err == nil {
			projects = append(projects, project)
			err = jsonDec.Decode(&project)
		}
		if err == io.EOF {
			err = nil
		}
		errC <- err
	}()

	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, "error running go list command")
	}
	if err := <-errC; err != nil {
		return nil, errors.Wrap(err, "error parsing go list output")
	}

	return projects, nil
}
