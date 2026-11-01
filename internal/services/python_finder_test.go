package services

import (
	"errors"
	"testing"
)

func fakeLookPath(found string) func(string) (string, error) {
	return func(string) (string, error) {
		if found == "" {
			return "", errors.New("not found")
		}
		return found, nil
	}
}

func installed(paths ...string) func(string) bool {
	return func(path string) bool {
		for _, candidate := range paths {
			if candidate == path {
				return true
			}
		}
		return false
	}
}

func TestOnMacThePythonOrgInstallWinsOverApplesStandIn(t *testing.T) {
	python := FindPython("darwin", installed("/Library/Frameworks/Python.framework/Versions/Current/bin/python3", "/opt/homebrew/bin/python3"), fakeLookPath("/usr/bin/python3"))
	if python != "/Library/Frameworks/Python.framework/Versions/Current/bin/python3" {
		t.Errorf("got %s", python)
	}
}

func TestOnMacHomebrewIsUsedWithoutThePythonOrgInstall(t *testing.T) {
	if python := FindPython("darwin", installed("/opt/homebrew/bin/python3"), fakeLookPath("/usr/bin/python3")); python != "/opt/homebrew/bin/python3" {
		t.Errorf("got %s", python)
	}
}

func TestOnMacWithoutAnInstallPathIsUsed(t *testing.T) {
	if python := FindPython("darwin", installed(), fakeLookPath("/usr/bin/python3")); python != "/usr/bin/python3" {
		t.Errorf("got %s", python)
	}
}

func TestOnLinuxPathIsUsed(t *testing.T) {
	if python := FindPython("linux", installed("/opt/homebrew/bin/python3"), fakeLookPath("/usr/bin/python3")); python != "/usr/bin/python3" {
		t.Errorf("got %s", python)
	}
}

func TestWithoutAnyPythonTheNameIsKeptForTheError(t *testing.T) {
	if python := FindPython("linux", installed(), fakeLookPath("")); python != "python3" {
		t.Errorf("got %s", python)
	}
}
