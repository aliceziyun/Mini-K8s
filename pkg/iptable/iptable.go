package iptable

import (
	"bytes"
	"io"
	"os/exec"
	"strconv"
)

type IPTable struct {
	path    string
	timeout int
}

func New() (*IPTable, error) {

	path, err := exec.LookPath("iptables")
	if err != nil {
		return nil, err
	}

	t := &IPTable{
		path:    path,
		timeout: 0,
	}

	return t, nil
}

// run runs an iptables command with the given arguments, ignoring
// any stdout output
func (t *IPTable) run(args ...string) error {
	return t.runWithOutput(args, nil)
}

// runWithOutput runs an iptables command with the given arguments,
// writing any stdout output to the given writer
func (t *IPTable) runWithOutput(args []string, stdout io.Writer) error {
	args = append([]string{t.path}, args...)
	//if t.hasWait {
	//	args = append(args, "--wait")
	//	if t.timeout != 0 && t.waitSupportSecond {
	//		args = append(args, strconv.Itoa(t.timeout))
	//	}
	//} else {
	//fmu, err := newXtablesFileLock()
	//if err != nil {
	//	return err
	//}
	//ul, err := fmu.tryLock()
	//if err != nil {
	//	syscall.Close(fmu.fd)
	//	return err
	//}
	//defer ul.Unlock()
	//}

	var stderr bytes.Buffer
	cmd := exec.Cmd{
		Path:   t.path,
		Args:   args,
		Stdout: stdout,
		Stderr: &stderr,
	}

	if err := cmd.Run(); err != nil {
		switch e := err.(type) {
		case *exec.ExitError:
			return &Error{*e, cmd, stderr.String(), nil}
		default:
			return err
		}
	}

	return nil
}

func (t *IPTable) IsChainExist(table string, chain string) (bool, error) {
	err := t.run("-t", table, "-S", chain, "1")
	err1, eok := err.(*Error)
	switch {
	case err == nil:
		return true, nil
	case eok && err1.ExitStatus() == 1:
		return false, nil
	default:
		return false, err
	}
}

func (t *IPTable) NewChain(table string, chain string) error {
	return t.run("-t", table, "-N", chain)
}

func (t *IPTable) Insert(table string, chain string, pos int, ruleSpec ...string) error {
	cmd := append([]string{"-t", table, "-I", chain, strconv.Itoa(pos)}, ruleSpec...)
	return t.run(cmd...)
}
