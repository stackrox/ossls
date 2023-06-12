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
}

func ProjectsFromGoModFile(filename string) ([]GoModProject, error) {
	cmd := exec.Command("go", "mod", "download", "-json")
	cmd.Dir = filepath.Dir(filename)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create pipe for redirecting stdout")
	}
	if err := cmd.Start(); err != nil {
		return nil, errors.Wrap(err, "error starting go list command")
	}

	var projects []GoModProject
	errC := make(chan error, 1)
	go func() {
		jsonDec := json.NewDecoder(stdout)
		var project GoModProject
		err := jsonDec.Decode(&project)
		for err == nil {
			if project.Dir != "" {
				projects = append(projects, project)
			}

			project = GoModProject{}
			err = jsonDec.Decode(&project)
		}
		if err == io.EOF {
			err = nil
		}
		errC <- err
	}()

	if err := <-errC; err != nil {
		return nil, errors.Wrap(err, "error parsing go list output")
	}
	if err := cmd.Wait(); err != nil {
		return nil, errors.Wrap(err, "error waiting for go list command")
	}

	return projects, nil
}
