package command

import (
	"os"
	"runtime/debug"
	"user-manager/util/errors"
	"user-manager/util/logger"
)

type Command func(log logger.Logger) error

func Run(topic string, command Command) {
	logger.SetLogJSON(os.Getenv("LOG_JSON") != "")
	log := logger.Log(topic)
	exitCode := 0

	defer func() {
		if p := recover(); p != nil {
			log.Err(errors.WrapRecoveredPanic(p, debug.Stack()))
			exitCode = 1
		}
		log.Info("Shutdown complete. ExitCode: %d", exitCode)
		os.Exit(exitCode)
	}()

	if err := command(log); err != nil {
		log.Err(errors.Wrap("command returned error", err))
		log.Warn("Exited abnormally")
		exitCode = 1
	}
}
