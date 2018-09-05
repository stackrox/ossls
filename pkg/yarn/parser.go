package yarn

import (
	"bufio"
	"os"
	"strings"
)

type kind int

const (
	Blank kind = iota
	Comment
	Spec
	Property
	Header
	Dep
)

func ParseScanner(scanner *bufio.Scanner) ([]Entry, error) {
	var (
		entries      []Entry
		lastHeader   string
		currentEntry *Entry
	)

	for scanner.Scan() {
		line := scanner.Text()

		switch lineType(line) {
		case Blank:
			continue

		case Comment:
			continue

		case Spec:
			name, versions := parseSpecLine(strings.TrimSuffix(strings.TrimSpace(line), ":"))

			if currentEntry != nil {
				entries = append(entries, *currentEntry)
			}

			currentEntry = &Entry{
				Name:  name,
				Specs: versions,
			}

		case Property:
			name, value := parsePropertyLine(strings.TrimSpace(line))
			addProp(currentEntry, name, value)

		case Header:
			lastHeader = strings.TrimSuffix(strings.TrimSpace(line), ":")

		case Dep:
			name, version := parsePropertyLine(strings.TrimSpace(line))
			addDep(currentEntry, lastHeader, name, version)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if currentEntry != nil {
		entries = append(entries, *currentEntry)
	}

	return entries, nil
}

func Parse(filename string) ([]Entry, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	return ParseScanner(scanner)
}

func lineType(line string) kind {
	switch {
	case strings.TrimSpace(line) == "":
		return Blank
	case strings.HasPrefix(strings.TrimSpace(line), "#"):
		return Comment
	case strings.HasSuffix(line, ":") && strings.HasPrefix(line, "  "):
		return Header
	case strings.HasSuffix(line, ":"):
		return Spec
	case strings.HasPrefix(line, "    "):
		return Dep
	case strings.HasPrefix(line, "  "):
		return Property
	}
	panic("unknown line type")
}

func addDep(entry *Entry, depType string, name string, version string) {
	switch depType {
	case "dependencies":
		entry.Dependencies = append(entry.Dependencies, Dependency{name, version})
	case "optionalDependencies":
		entry.OptionalDependencies = append(entry.OptionalDependencies, Dependency{name, version})
	default:
		panic("unknown dependency type")
	}
}

func addProp(entry *Entry, name string, value string) {
	switch name {
	case "version":
		entry.Version = value
	case "resolved":
		entry.Resolved = value
	default:
		panic("unknown property type")
	}
}

func parseSpecLine(line string) (string, []string) {
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, ":")
	var (
		parts    = strings.Split(line, ",")
		versions = make([]string, 0)
		name     string
	)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		part = strings.Trim(part, `"`)

		pieces := strings.Split(part, "@")
		var version string
		switch len(pieces) {
		case 2:
			if name != "" && name != pieces[0] {
				panic("spec name changed unexpectedly")
			}
			name = strings.TrimSpace(pieces[0])
			version = strings.TrimSpace(pieces[1])
		case 3:
			if name != "" && name != "@"+pieces[1] {
				panic("spec name changed unexpectedly")
			}
			name = "@" + strings.TrimSpace(pieces[1])
			version = strings.TrimSpace(pieces[2])
		}

		versions = append(versions, version)
	}

	return name, versions
}

func parsePropertyLine(line string) (string, string) {
	line = strings.TrimSpace(line)
	parts := strings.SplitN(line, " ", 2)

	key := parts[0]
	key = strings.TrimSpace(key)
	key = strings.Trim(key, `"`)

	value := parts[1]
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"`)

	return key, value
}
