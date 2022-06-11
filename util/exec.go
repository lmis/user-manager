package util

import (
	"os/exec"
)

func RunShellCommand(command string) error {
	log := Log("SHELL_COMMAND")
	log.Info("$ %s", command)
	cmd := exec.Command("sh", "-c", command)

	out, err := cmd.CombinedOutput()
	log.Info("| %s", string(out))

	return err
}
