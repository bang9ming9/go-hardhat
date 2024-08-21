package main

import (
	"fmt"
	"os"

	"github.com/bang9ming9/go-hardhat/internal/compile"
	initCommand "github.com/bang9ming9/go-hardhat/internal/init"
	"github.com/urfave/cli/v2"
)

var (
	app = cli.NewApp()
)

func init() {
	app.Name = "GO-HARDHAT"
	app.Copyright = "bang9ming9"
	app.CommandNotFound = func(ctx *cli.Context, s string) {
		cli.ShowAppHelp(ctx)
		os.Exit(1)
	}
	app.OnUsageError = func(ctx *cli.Context, err error, isSubcommand bool) error {
		cli.ShowAppHelp(ctx)
		return err
	}

	app.Commands = append(app.Commands, []*cli.Command{
		initCommand.Command,
		compile.Command,
	}...)
}

func main() {
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
