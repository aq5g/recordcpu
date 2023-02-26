package scmd

import (
	"bytes"
	"os/exec"
)

func RunCmdwithPipe(cmdStr string) (string, error) {
	cmd := exec.Command("bash", "-c", cmdStr)
	var out bytes.Buffer
	var oerr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &oerr
	err := cmd.Run()
	if err != nil {
		return oerr.String(), err
	} else {
		return out.String(), nil
	}
}
