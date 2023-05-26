package commands

import (
	"os"

	"github.com/urfave/cli"
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
