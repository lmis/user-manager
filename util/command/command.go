package command

import (
	"os"
	"path/filepath"
	"runtime/debug"
	"user-manager/util/errors"
	"user-manager/util/logger"
)

type Command func(log logger.Logger, dir string) error

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

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Err(errors.Wrap("cannot get executable path", err))
	}

	if err := command(log, dir); err != nil {
		log.Err(errors.Wrap("command returned error", err))
		log.Warn("Server exited abnormally")
		exitCode = 1
	}
}
