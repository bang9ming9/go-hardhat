package utils

import (
	"fmt"
	"regexp"

	"github.com/fabelx/go-solc-select/pkg/installer"
	"github.com/fabelx/go-solc-select/pkg/versions"
)

func VerifySolidityVersion(version string) error {
	reg := regexp.MustCompile(`^\d+(\.\d+){2}$`)
	if !reg.MatchString(version) {
		return fmt.Errorf("%s is invalid solidity version", version)
	}
	return nil
}

func InstallSolc(version string) error {
	if err := VerifySolidityVersion(version); err != nil {
		return err
	}

	// 이미 다운이 되어있는경우
	if _, ok := versions.GetInstalled()[version]; ok {
		return nil
	}

	return installer.InstallSolc(version)
}
