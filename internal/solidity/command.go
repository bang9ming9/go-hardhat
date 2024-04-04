package solidity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bang9ming9/go-hardhat/internal/utils"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

var Command *cli.Command = &cli.Command{
	Name:   "setsolc",
	Action: command,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "version",
		},
		&cli.StringFlag{
			Name: "remappings",
		},
	},
}

func command(ctx *cli.Context) error {
	config, err := utils.ReadConfig()
	if err != nil {
		return err
	}

	if ctx.IsSet("version") {
		version := ctx.String("version")
		if err := utils.InstallSolc(version); err != nil {
			return err
		}
		config.Set(utils.SOLC_VERSION, version)
	}

	if ctx.IsSet("remappings") {
		remappings, err := appendRemappings(config.GetStringSlice(utils.SOLC_REMAPPINGS), strings.Split(ctx.String("remappings"), ","))
		if err != nil {
			return err
		}
		config.Set(utils.SOLC_REMAPPINGS, remappings)
	}

	return config.WriteConfig()
}

func appendRemappings(remappings []string, newMappings []string) ([]string, error) {
	for _, newmapping := range newMappings {
		key, remapping, err := parseKeyPath(newmapping)
		if err != nil {
			return nil, err
		}
		// 중복 확인
		var isNew = true
		for i, mapping := range remappings {
			if strings.HasPrefix(mapping, key) {
				isNew = false
				if mapping != remapping {
					if ok, err := prompt.Stdin.PromptConfirm(fmt.Sprintf("do you want chante %s to %s? ", mapping, remapping)); err != nil {
						return nil, errors.Wrap(err, "PromptConfirm")
					} else if ok {
						remappings[i] = remapping
						break
					}
				}
			}
		}
		if isNew {
			remappings = append(remappings, remapping)
		}
	}

	return remappings, nil
}

func parseKeyPath(mapping string) (string, string, error) {
	pathSeparator := string(os.PathSeparator)
	s := strings.Split(mapping, "=")
	if len(s) != 2 {
		return "", "", fmt.Errorf("%s is invalid format", mapping)
	}
	key, path := s[0], s[1]

	if !strings.HasPrefix(key, "@") {
		return "", "", fmt.Errorf("%s is invalid format (please start with '@')", mapping)
	}

	if !strings.HasSuffix(key, pathSeparator) {
		key += pathSeparator
	}

	if !filepath.IsAbs(path) {
		return "", "", fmt.Errorf("%s is invalid format (please absolute path)", path)
	}
	if _, err := os.Stat(path); err != nil {
		return "", "", errors.Wrap(err, "os.Stat "+path)
	}

	if !strings.HasSuffix(path, pathSeparator) {
		path += pathSeparator
	}

	return key, fmt.Sprintf("%s=%s", key, path), nil
}
