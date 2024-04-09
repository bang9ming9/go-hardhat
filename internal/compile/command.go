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
	// 1-2. 컴파일한 파일을 임시로 저장할 디렉토리를 만든다. (command 종료와 동시에 제거)
	if err := os.MkdirAll(utils.GetABIsCompiledDir(), 0755); err != nil {
		return errors.Wrap(err, "os.MkdirAll")
	}
	defer os.RemoveAll(utils.GetABIsCompiledDir())

	// 2. solc-0.0.0 --optimize --combined-json abi,bin contracts/*.sol
	contracts, err := compile(version, remappings)
	if err != nil {
		return errors.Wrap(err, "compile")
	}

	// abigen 실행
	for _, name := range contracts {
		if err := abigen(name); err != nil {
			return errors.Wrap(err, name)
		}
	}
	if ctx.Bool("tidy") {
		return exec.Command("go", "mod", "tidy").Run()
	} else {
		return nil
	}
}

func compile(version string, remappings []string) ([]string, error) {
	// ~/.gsolc-select/artifacts/solc-a.b.c/solc-a.b.c
	cmdName := filepath.Join(solconfig.SolcArtifacts, fmt.Sprintf("solc-%s", version), fmt.Sprintf("solc-%s", version))
	args := append([]string{"--optimize", "--combined-json", "bin,abi"}, remappings...)

	dirpath := utils.GetContractDir()
	files, err := findAllSolidityFiles(dirpath)
	if err != nil {
		return nil, errors.Wrap(err, "findAllSolidityFiles")
	}

	contracts := make(map[string]struct{}, 0)
	abiscompiledDir := utils.GetABIsCompiledDir()
	for _, name := range files {
		cmd := exec.Command(cmdName, append(args, name)...)
		var stderr, stdout bytes.Buffer
		cmd.Stderr, cmd.Stdout = &stderr, &stdout
		if err := cmd.Run(); err != nil {
			return nil, errors.Wrap(err, stderr.String())
		}

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
					if abibytes, err := json.Marshal(value.ABI); err != nil {
						return nil, errors.Wrap(err, "marshal abi")
					} else if err := os.WriteFile(
						filepath.Join(abiscompiledDir, contract)+".abi",
						abibytes,
						0644,
					); err != nil {
						return nil, errors.Wrap(err, "write abi")
					} else if err := os.WriteFile(
						filepath.Join(abiscompiledDir, contract)+".bin",
						[]byte("0x"+value.BIN),
						0644,
					); err != nil {
						return nil, errors.Wrap(err, "write bin")
					} else {
						contracts[contract] = struct{}{}
					}
				}
			}
		}
	}

	keys := make([]string, 0, len(contracts))
	for k := range contracts {
		keys = append(keys, k)
	}

	return keys, nil
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

func abigen(name string) error {
	compiledDir := utils.GetABIsCompiledDir()
	out := utils.GetABIsDir()

	cmd := exec.Command("abigen",
		"--pkg", "abis",
		"--out", filepath.Join(out, name+".go"),
		"--abi", filepath.Join(compiledDir, name+".abi"),
		"--bin", filepath.Join(compiledDir, name+".bin"),
		"--type", name,
	)
	var stderr, stdout bytes.Buffer
	cmd.Stderr, cmd.Stdout = &stderr, &stdout

	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, stderr.String())
	}
	return nil
}
