package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/fabelx/go-solc-select/pkg/installer"
	"github.com/fabelx/go-solc-select/pkg/versions"
	"github.com/pkg/errors"
)

func InstallSolc(version string) error {
	var err error
	if version, err = ToSolcVersion(version); err != nil {
		return err
	}

	// 이미 다운이 되어있는경우
	if _, ok := versions.GetInstalled()[version]; ok {
		return nil
	}

	return installer.InstallSolc(version)
}

func ToSolcVersion(version string) (string, error) {
	if version == "" {
		cmd := exec.Command("solc", "--version")
		var stderr, stdout bytes.Buffer
		cmd.Stderr, cmd.Stdout = &stderr, &stdout

		if err := cmd.Run(); err != nil {
			return "", errors.Wrap(err, stderr.String())
		}
		version = stdout.String()
	}
	reg := regexp.MustCompile(`\d+\.\d+\.\d+`)
	match := reg.FindString(version)
	if match == "" {
		return "", fmt.Errorf("%s is invalid solidity version", version)
	}
	return match, nil
}
