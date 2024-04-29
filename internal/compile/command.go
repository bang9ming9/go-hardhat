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
	"github.com/ethereum/go-ethereum/console/prompt"
	solconfig "github.com/fabelx/go-solc-select/pkg/config"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var Command *cli.Command = &cli.Command{
	Name:   "compile",
	Action: command,
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
			Name:  "exc",
			Usage: "Comma separated types to exclude from binding",
		}, &cli.StringFlag{
			Name:  "filter",
			Usage: "Comma separated types to filter from binding",
		},
	},
}

func command(ctx *cli.Context) error {
	config, err := utils.ReadConfig()
	if err != nil {
		return errors.Wrap(err, "utils.ReadConfig")
	}
	defer config.WriteConfig()
	// solc 확인 및 다운
	var version string = ""
	if config.IsSet(utils.SOLC_VERSION) {
		version = config.GetString(utils.SOLC_VERSION)
	} else {
		if input, err := prompt.Stdin.PromptInput("solidity version: "); err != nil {
			return err
		} else {
			version = input
		}
	}
	if err := utils.InstallSolc(version); err != nil {
		return errors.Wrap(err, "utils.InstallSolc")
	}
	config.Set(utils.SOLC_VERSION, version)

	// compile 실행
	// 1. remappings 셋팅
	var remappings []string
	if config.IsSet(utils.SOLC_REMAPPINGS) {
		remappings = config.GetStringSlice(utils.SOLC_REMAPPINGS)
	}
	// 1-1. remappings openzeppelin
	ozPath := filepath.Join(utils.GetNodeModulesDir(), "@openzeppelin")
	if _, err := os.Stat(ozPath); err == nil {
		var ok = false
		for _, line := range remappings {
			if strings.HasPrefix(line, "@openzeppelin") {
				ok = true
				break
			}
		}
		if !ok {
			remappings = append(remappings, fmt.Sprintf("@openzeppelin/=%s/", ozPath))
			config.Set(utils.SOLC_REMAPPINGS, remappings)
		}
	}

	// 2. solc-0.0.0 --optimize --combined-json abi,bin contracts/*.sol
	contracts, err := compile(version, remappings)
	if err != nil {
		return errors.Wrap(err, "compile")
	}

	// 3. filter, exclude 적용
	if ctx.IsSet("filter") && ctx.IsSet("exc") {
		return errors.New("Flags 'filter' and 'exc' can't be used at the same time")
	} else if ctx.IsSet("filter") {
		filters := make(map[string]struct{})
		for _, f := range strings.Split(ctx.String("filter"), ",") {
			filters[f] = struct{}{}
		}
		for name := range contracts {
			if _, ok := filters[name]; !ok {
				delete(contracts, name)
			}
		}
	} else if ctx.IsSet("exc") {
		for _, exc := range strings.Split(ctx.String("exc"), ",") {
			delete(contracts, exc)
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
}

type compiled struct {
	ABI string `json:"abi"`
	BIN string `json:"bin"`
}

func compile(version string, remappings []string) (map[string]compiled, error) {
	// ~/.gsolc-select/artifacts/solc-a.b.c/solc-a.b.c
	cmdName := filepath.Join(solconfig.SolcArtifacts, fmt.Sprintf("solc-%s", version), fmt.Sprintf("solc-%s", version))
	args := append([]string{"--optimize", "--combined-json", "bin,abi"}, remappings...)
	args = append(args, "--")

	dirpath := utils.GetContractDir()
	files, err := findAllSolidityFiles(dirpath)
	if err != nil {
		return nil, errors.Wrap(err, "findAllSolidityFiles")
	}

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
		path, err := filepath.Abs(s[0])
		if err != nil {
			return nil, err
		}
		if strings.HasPrefix(path, dirpath) {
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
	}

	return contracts, nil
}

func findAllSolidityFiles(rootDir string) ([]string, error) {
	var solFiles []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			dir := info.Name()
			if len(dir) > len(rootDir) && strings.HasPrefix(dir, rootDir) {
				if inDir, err := findAllSolidityFiles(info.Name()); err != nil {
					return err
				} else if inDir != nil {
					solFiles = append(solFiles, inDir...)
				}
			}
		} else if strings.HasSuffix(info.Name(), ".sol") {
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
