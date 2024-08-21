package compile

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bang9ming9/go-hardhat/internal/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	solconfig "github.com/fabelx/go-solc-select/pkg/config"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var Command *cli.Command = &cli.Command{
	Name:      "compile",
	ArgsUsage: "<solc-version>",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "tidy",
			Value: false,
			Usage: "execute with go mod tidy",
		}, &cli.BoolFlag{
			Name:  "merge",
			Value: false,
			Usage: "write all bind codes to abis/bind.go",
		}, &cli.StringFlag{
			Name:    "exclude",
			Aliases: []string{"e", "exc"},
			Usage:   "Comma separated path to exclude from compile",
		}, &cli.StringFlag{
			Name:    "filter",
			Aliases: []string{"f"},
			Usage:   "Comma separated types to filter from binding",
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := utils.SetDirPath(); err != nil {
			return nil
		}

		// 1. solc 버전 확인 및 설치
		version, err := utils.ToSolcVersion(ctx.Args().First())
		if err != nil {
			return errors.Wrap(err, "utils.ToSolcVersion")
		}
		if err := utils.InstallSolc(version); err != nil {
			return errors.Wrap(err, "utils.InstallSolc")
		}

		// 2. solidity 컴파일 실행
		// 2-1 컴파일 할 파일 목록 가져오기
		files, err := findSolFiles(utils.GetContractDir(), strings.Split(ctx.String("exclude"), ","))
		if err != nil {
			return errors.Wrap(err, "findSolFiles")
		}

		// 2-2. compile 실행 (solc-0.0.0 --optimize --combined-json abi,bin contracts/*.sol)
		contracts, err := compile(version, utils.ReadRemappings(), files)
		if err != nil {
			return errors.Wrap(err, "compile")
		}

		// 3. filter 적용
		if ctx.IsSet("filter") {
			filters := make(map[string]struct{})
			for _, f := range strings.Split(ctx.String("filter"), ",") {
				filters[f] = struct{}{}
			}
			for name := range contracts {
				if _, ok := filters[name]; !ok {
					delete(contracts, name)
				}
			}
		}

		// 4. abigen 실행
		abisDir := utils.GetABIsDir()
		if _, err := os.Stat(abisDir); errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(abisDir, 0755); err != nil {
				return errors.Wrap(err, abisDir)
			}
		}

		if ctx.Bool("merge") {
			if err := abigenMerge(contracts); err != nil {
				return errors.Wrap(err, "abigenMerge")
			}
		} else {
			for name, compiled := range contracts {
				if err := abigen(name, compiled); err != nil {
					return errors.Wrap(err, name)
				}
			}
		}

		if ctx.Bool("tidy") {
			return exec.Command("go", "mod", "tidy").Run()
		} else {
			return nil
		}
	},
}

type compiled struct {
	ABI string `json:"abi"`
	BIN string `json:"bin"`
}

func compile(version string, remappings []string, files []string) (map[string]compiled, error) {
	// ~/.gsolc-select/artifacts/solc-a.b.c/solc-a.b.c
	cmdName := filepath.Join(solconfig.SolcArtifacts, fmt.Sprintf("solc-%s", version), fmt.Sprintf("solc-%s", version))
	args := append([]string{"--optimize", "--combined-json", "bin,abi"}, remappings...)
	args = append(args, "--")

	cmd := exec.Command(cmdName, append(args, files...)...)
	var stderr, stdout bytes.Buffer
	cmd.Stderr, cmd.Stdout = &stderr, &stdout
	if err := cmd.Run(); err != nil {
		return nil, errors.Wrap(err, stderr.String())
	}

	contracts := make(map[string]compiled)
	var output = struct {
		Contracts map[string]struct {
			ABI interface{} `json:"abi"`
			BIN string      `json:"bin"`
		} `json:"contracts"`
		Version string `json:"version"`
	}{}

	if err := json.Unmarshal(stdout.Bytes(), &output); err != nil {
		return nil, errors.Wrap(err, "json.Unmarshal")
	}

	for pathname, value := range output.Contracts {
		s := strings.Split(pathname, ":")

		contract := s[1]
		if _, ok := contracts[contract]; ok {
			return nil, fmt.Errorf("%s is duplicated", contract)
		}
		if value.BIN != "" { // interface
			abiString, ok := value.ABI.(string)
			if !ok {
				abibytes, err := json.Marshal(value.ABI)
				if err != nil {
					return nil, errors.Wrap(err, "marshal abi")
				}
				abiString = string(abibytes)
			}
			if abiString == "[]" { // abstract contract, or library
				continue
			}
			contracts[contract] = compiled{
				ABI: abiString,
				BIN: "0x" + value.BIN,
			}
		}
	}

	return contracts, nil
}

func findSolFiles(rootDir string, excludes []string) ([]string, error) {
	var solFiles []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".sol" {
			for _, exc := range excludes {
				if strings.HasPrefix(path, exc) {
					return nil
				}
			}
			solFiles = append(solFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return solFiles, nil
}

func abigen(name string, contract compiled) error {
	str, err := bind.Bind(
		[]string{name},
		[]string{contract.ABI},
		[]string{contract.BIN},
		nil,
		"abis",
		bind.LangGo,
		nil, nil,
	)
	if err != nil {
		return errors.Wrap(err, name)
	}

	return os.WriteFile(filepath.Join(utils.GetABIsDir(), name+".go"), []byte(str), 0600)
}

func abigenMerge(contracts map[string]compiled) error {
	var types, abis, bytecodes []string = make([]string, 0), make([]string, 0), make([]string, 0)
	for name, compiled := range contracts {
		types = append(types, name)
		abis = append(abis, compiled.ABI)
		bytecodes = append(bytecodes, compiled.BIN)
	}
	str, err := bind.Bind(
		types,
		abis,
		bytecodes,
		nil,
		"abis",
		bind.LangGo,
		nil, nil,
	)
	if err != nil {
		return errors.Wrap(err, "abigenMerge")
	}

	return os.WriteFile(filepath.Join(utils.GetABIsDir(), "bind.go"), []byte(str), 0600)
}
