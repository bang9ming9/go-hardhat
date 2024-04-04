package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var (
	rootpath     string = ""
	config       string = ""
	contract     string = ""
	test         string = ""
	nodemudules  string = ""
	abis         string = ""
	abisCompiled string = ""
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
	config = filepath.Join(rootpath, ".config")
	contract = filepath.Join(rootpath, "contracts")
	test = filepath.Join(rootpath, "test")
	nodemudules = filepath.Join(rootpath, "node_modules")
	abis = filepath.Join(rootpath, "abis")
	abisCompiled = filepath.Join(abis, "compiled")
	return nil
}

func GetConfigtDir() string {
	return config
}

func GetContractDir() string {
	return contract
}

func GetTestDir() string {
	return test
}

func GetNodeModulesDir() string {
	return nodemudules
}

func GetABIsDir() string {
	return abis
}

func GetABIsCompiledDir() string {
	return abisCompiled
}

func ReadConfig() (*viper.Viper, error) {
	if err := SetDirPath(); err != nil {
		return nil, err
	}
	v := viper.New()
	v.SetConfigType("toml")
	v.SetConfigName("config.toml")
	v.AddConfigPath(GetConfigtDir())
	return v, v.ReadInConfig()
}
