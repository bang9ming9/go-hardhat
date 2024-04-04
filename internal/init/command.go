package init

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bang9ming9/go-hardhat/internal/utils"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var Command *cli.Command = &cli.Command{
	Name:   "init",
	Action: command,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "modulepath",
		},
	},
}

func command(cCtx *cli.Context) error {
	rootpath, err := utils.GetRootPath()
	if err != nil {
		return err
	}

	// go.mod 이 없다면 생성 한다.
	if rootpath == "" {
		if err := initGoModule(cCtx.String("modulepath")); err != nil {
			return err
		}
	}

	if err := utils.SetDirPath(); err != nil {
		return errors.Wrap(err, "utils.SetDirPath")
	}

	// 기본 디렉토리, 파일을 생성한다.
	if err := makeDefaultFS(); err != nil {
		return err
	}

	if err := doNPMDependency(); err != nil {
		return err
	}

	return nil
}

func initGoModule(modulepath string) error {
	if modulepath == "" {
		if path, err := prompt.Stdin.PromptInput("Go module path:"); err != nil {
			return nil
		} else {
			modulepath = strings.TrimSpace(path)
		}
	}

	if !regexp.MustCompile("^[A-Za-z0-9\"-/]+$").MatchString(modulepath) {
		return fmt.Errorf("%s is invalid go mod path", modulepath)
	}

	err := exec.Command("go", "mod", "init", modulepath).Run()
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("go mod init %s", modulepath))
	}
	return err
}

func makeDefaultFS() error {
	configDir := utils.GetConfigtDir()
	if _, err := os.Stat(configDir); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(configDir, 0755); err != nil {
			return errors.Wrap(err, configDir)
		}
	}

	configFile := filepath.Join(configDir, "config.toml")
	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(configFile, []byte{}, 0644); err != nil {
			return errors.Wrap(err, configFile)
		}
	}

	contractsDir := utils.GetContractDir()
	if _, err := os.Stat(contractsDir); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(contractsDir, 0755); err != nil {
			return errors.Wrap(err, contractsDir)
		}
	}

	testDir := utils.GetTestDir()
	if _, err := os.Stat(testDir); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(testDir, 0755); err != nil {
			return errors.Wrap(err, testDir)
		}
	}

	return nil
}

func doNPMDependency() error {
	rootpath, err := utils.GetRootPath()
	if err != nil {
		return err
	}
	if err := exec.Command("cd", rootpath).Run(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("cd %s", rootpath))
	}
	if err := exec.Command("npm", "install", "--save-dev", "prettier", "prettier-plugin-solidity").Run(); err != nil {
		return errors.Wrap(err, "npm install prettier")
	}
	prettierrc := filepath.Join(rootpath, ".prettierrc")
	if _, err := os.Stat(prettierrc); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(prettierrc, []byte(`
{
	"overrides": [
		{
		"files": "*.sol",
		"options": {
			"printWidth": 80,
			"tabWidth": 4,
			"useTabs": true,
			"singleQuote": false,
			"bracketSpacing": true,
			"explicitTypes": "always"
		}
		}
	]
}
		`), 0644); err != nil {
			return errors.Wrap(err, fmt.Sprintf("create %s", prettierrc))
		}
	}

	return nil
}
