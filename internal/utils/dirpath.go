package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

var (
	rootpath       string = ""
	contract       string = ""
	test           string = ""
	abis           string = ""
	remappingspath string = ""
)

func GetRootPath() (string, error) {
	if rootpath == "" {
		cmd := exec.Command("go", "env")
		output, err := cmd.Output()
		if err != nil {
			return "", errors.Wrap(err, "cmd.Output(go mod)")
		}
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "GOMOD=") {
				rootpath = filepath.Dir(strings.TrimPrefix(line, "GOMOD='"))
				if !strings.HasPrefix(rootpath, os.Getenv("HOME")) {
					rootpath = ""
				}
				break
			}
		}
	}
	return rootpath, nil
}

func SetDirPath() error {
	rootpath, err := GetRootPath()
	if err != nil {
		return errors.Wrap(err, "GetRootPath")
	}
	if _, err := os.Stat(rootpath); err != nil {
		return errors.Wrap(err, "os.Stat")
	}
	contract = filepath.Join(rootpath, "contracts")
	test = filepath.Join(rootpath, "test")
	abis = filepath.Join(rootpath, "abis")
	remappingspath = filepath.Join(contract, "remappings.txt")
	return nil
}

func GetContractDir() string {
	return contract
}

func GetTestDir() string {
	return test
}

func GetABIsDir() string {
	return abis
}

func GetRemappingsFilePath() string {
	return remappingspath
}

func ReadRemappings() []string {
	if remappingspath == "" {
		return nil
	}

	bytes, err := os.ReadFile(remappingspath)
	if err != nil {
		return nil
	}

	return strings.Split(string(bytes), "\n")
}
