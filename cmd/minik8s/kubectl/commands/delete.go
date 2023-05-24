package commands

import (
	_const "Mini-K8s/cmd/const"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/urfave/cli"
	"net/http"
	"strings"
)

func NewDeleteCommand() cli.Command {
	deleteCmd := cli.Command{
		Name:  "delete",
		Usage: "delete <resource> <resourceName>",
		Action: func(c *cli.Context) error {
			if len(c.Args()) == 2 {
				deleteResource(c.Args().Get(0), c.Args().Get(1))
			} else {
				fmt.Println("You need to specify a resource object and resource Name!")
				fmt.Printf("[Possible Resource Object]: ")
				printPossibleResourceObj()
			}
			return nil
		},
	}
	return deleteCmd
}

func deleteResource(resource string, name string) {
	switch resource {
	case strings.ToLower(resourceList[0]):
		url := _const.BASE_URI + _const.POD_CONFIG_PREFIX
		err := deleteObj(url, name)
		if err != nil {
			fmt.Println(err)
			return
		}
		break
	default:
		fmt.Println("No such resource!")
		fmt.Printf("[Possible Resource Object]: ")
		printPossibleResourceObj()
		return
	}
}

func deleteObj(url string, name string) error {
	nameRaw, _ := json.Marshal(name)
	reqBody := bytes.NewBuffer(nameRaw)
	request, err := http.NewRequest("DELETE", url, reqBody)
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return errors.New("StatusCode not 200")
	}
	return nil
}
