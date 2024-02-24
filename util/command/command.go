package command

import (
	"fmt"
	"os"
	"runtime/debug"
	"user-manager/util/errs"
	"user-manager/util/logger"
)

type Command func(log logger.Logger) error

func Run(topic string, command Command) {
	logger.SetLogJSON(os.Getenv("LOG_JSON") != "")
	log := logger.NewLogger(topic)
	exitCode := 0

	defer func() {
		if p := recover(); p != nil {
			log.Err(errs.WrapRecoveredPanic(p, debug.Stack()))
			exitCode = 1
		}
		log.Info(fmt.Sprintf("Shutdown complete. ExitCode: %d", exitCode))
		os.Exit(exitCode)
	}()

	if err := command(log); err != nil {
		log.Err(errs.Wrap("command returned error", err))
		log.Warn("Exited abnormally")
		exitCode = 1
	}
}
