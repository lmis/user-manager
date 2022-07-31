package util

import (
	"os"
	"path/filepath"
	"runtime/debug"
)

type Command func(log Logger, dir string) error

func Run(topic string, command Command) {
	SetLogJSON(os.Getenv("LOG_JSON") != "")
	log := Log(topic)
	exitCode := 0

	defer func() {
		if p := recover(); p != nil {
			log.Err(WrapRecoveredPanic(p, debug.Stack()))
			exitCode = 1
		}
		log.Info("Shutdown complete. ExitCode: %d", exitCode)
		os.Exit(exitCode)
	}()

	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Err(Wrap("cannot get executable path", err))
	}

	if err := command(log, dir); err != nil {
		log.Err(Wrap("command returned error", err))
		log.Warn("Server exited abnormally")
		exitCode = 1
	}
}
