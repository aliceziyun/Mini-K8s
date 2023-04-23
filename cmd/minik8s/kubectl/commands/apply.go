package commands

import (
	"fmt"
	"github.com/urfave/cli"
)

func NewApplyCommand() cli.Command {
	applyCmd := cli.Command{
		Name:  "apply",
		Usage: "create pod according to file",
		Action: func(c *cli.Context) error {
			fmt.Println("apply okk")
			return nil
		},
	}
	return applyCmd
}
