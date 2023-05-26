package object

import (
	_const "Mini-K8s/cmd/const"
	"errors"
	"fmt"
)

type Account struct {
	username       string
	password       string
	host           string
	remoteBasePath string
}

func NewAccountWith2Para(username string, password string) *Account {
	return &Account{
		username: username,
		password: password,
	}
}

func NewAccountWith4Para(username string, password string, host string, remoteBasePath string) *Account {
	return &Account{
		username:       username,
		password:       password,
		host:           host,
		remoteBasePath: remoteBasePath,
	}
}

func (account *Account) GetUsername() string {
	return account.username
}

func (account *Account) GetPassword() string {
	return account.password
}

func (account *Account) GetHost() string {
	return account.host
}

func (account *Account) GetRemoteBasePath() string {
	return account.remoteBasePath
}

func (account *Account) SetRemoteBasePath(host string) error {
	switch host {
	case _const.HostARM, _const.HostPiAndAI:
		account.host = host
		account.remoteBasePath = fmt.Sprintf("/lustre/home/acct-stu/%s", account.username)
		return nil
	case _const.HostSiyuan:
		account.host = host
		account.remoteBasePath = fmt.Sprintf("/dssg/home/acct-stu/%s", account.username)
		return nil
	default:
		return errors.New("unknown host type")
	}
}
