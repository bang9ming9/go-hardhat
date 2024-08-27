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
	Name: "init",
	Action: func(ctx *cli.Context) error {
		if err := initGoModule(ctx.Args().First()); err != nil {
			return errors.Wrap(err, "fail to init project")
		}

		if err := utils.SetDirPath(); err != nil {
			return errors.Wrap(err, "utils.SetDirPath")
		}

		// 기본 디렉토리, 파일을 생성한다.
		if err := makeDefaultFS(); err != nil {
			return err
		}

		return exec.Command("go", "mod", "tidy").Run()
	},
	ArgsUsage: "<module-path>",
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

	bmspath := fmt.Sprintf("github.com/bang9ming9/go-hardhat@%s", utils.AppVersion)
	if err := exec.Command("go", "get", bmspath).Run(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("go get %s", modulepath))
	}

	ethereumpath := fmt.Sprintf("github.com/ethereum/go-ethereum@%s", utils.EthereumVersion)
	if err := exec.Command("go", "get", ethereumpath).Run(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("go get %s", modulepath))
	}

	return nil
}

func makeDefaultFS() error {
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

	remapping := utils.GetRemappingsFilePath()
	if _, err := os.Stat(remapping); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(remapping, []byte{}, 0644); err != nil {
			return errors.Wrap(err, fmt.Sprintf("create %s", remapping))
		}
	}

	prettierrc := filepath.Join(contractsDir, ".prettierrc")
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
}`), 0644); err != nil {
			return errors.Wrap(err, fmt.Sprintf("create %s", prettierrc))
		}
	}

	basetest := filepath.Join(testDir, "base.go")
	if _, err := os.Stat(basetest); errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile(basetest, []byte(`package test

import (
	"github.com/bang9ming9/go-hardhat/bms"
	butils "github.com/bang9ming9/go-hardhat/bms/utils"
	"github.com/ethereum/go-ethereum/common"
)

var (
	_ = common.Big1
	_ = bms.ChainID
	_ = butils.ErrEventNotFind
)
`), 0644); err != nil {
			return errors.Wrap(err, fmt.Sprintf("create %s", basetest))
		}
	}

	return nil
}
