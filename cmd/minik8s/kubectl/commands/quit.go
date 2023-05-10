package commands

import (
	"github.com/urfave/cli"
	"os"
)

func NewQuitCommand() cli.Command {
	quitCmd := cli.Command{
		Name:  "quit",
		Usage: "quit the system",
		Action: func(c *cli.Context) error {
			os.Exit(0)
			return nil
		},
	}
	return quitCmd
}
