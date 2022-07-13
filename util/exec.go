package util

import (
	"os/exec"
)

func RunShellCommand(command string) error {
	log := Log("SHELL_COMMAND")
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
		return Wrap(message, err)
	}

	return nil
}
