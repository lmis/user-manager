package exec

import (
	"os/exec"
	"user-manager/util/errors"
	"user-manager/util/logger"
)

func RunShellCommand(command string) error {
	log := logger.Log("SHELL_COMMAND")
	log.Info("$ %s", command)
	cmd := exec.Command("sh", "-c", command)

	out, err := cmd.Output()
	outString := string(out)
	if outString != "" {
		log.Info("| %s", outString)
	}

	if err != nil {
		message := "error in executed command"
		if e, ok := err.(*exec.ExitError); ok {
			message += " (" + string(e.Stderr) + ")"
		}
		return errors.Wrap(message, err)
	}

	return nil
}
