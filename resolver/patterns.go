package resolver

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/stackrox/ossls/config"
)

type Matcher func(string) bool

type complexMatcher struct {
	include, exclude []Matcher
}

func (m complexMatcher) Match(name string) bool {
	for _, excludeMatcher := range m.exclude {
		if excludeMatcher(name) {
			return false
		}
	}

	for _, includeMatcher := range m.include {
		if includeMatcher(name) {
			return true
		}
	}

	return false
}

func CompilePatternConfig(cfg config.PatternConfig) (Matcher, error) {
	includes := make([]Matcher, 0, len(cfg.Patterns))
	for _, pat := range cfg.Patterns {
		matcher, err := compilePattern(pat)
		if err != nil {
			return nil, errors.Wrapf(err, "compiling pattern %q", pat)
		}
		includes = append(includes, matcher)
	}

	excludes := make([]Matcher, 0, len(cfg.ExcludePatterns))
	for _, pat := range cfg.ExcludePatterns {
		matcher, err := compilePattern(pat)
		if err != nil {
			return nil, errors.Wrapf(err, "compiling exclude pattern %q", pat)
		}
		excludes = append(excludes, matcher)
	}

	cm := complexMatcher{
		include: includes,
		exclude: excludes,
	}
	return cm.Match, nil
}

func compilePattern(pattern string) (Matcher, error) {
	if strings.HasPrefix(pattern, "~") {
		regexStr := strings.TrimPrefix(pattern, "~")
		regexMatcher, err := regexp.Compile(regexStr)
		if err != nil {
			return nil, errors.Wrap(err, "invalid regex pattern")
		}
		return regexMatcher.MatchString, nil
	}

	_, err := filepath.Match(pattern, "")
	if err != nil {
		return nil, errors.Wrap(err, "invalid glob pattern")
	}

	return func(name string) bool {
		matched, err := filepath.Match(pattern, name)
		if err != nil {
			panic(errors.Wrapf(err, "unexpected error matching pattern %q", pattern))
		}
		return matched
	}, nil
}
