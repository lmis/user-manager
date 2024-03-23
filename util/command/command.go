package command

import (
	"log/slog"
	"os"
	"runtime/debug"
	"user-manager/util/errs"
)

type Command func() error

func Run(command Command) {
	exitCode := 0

	defer func() {
		if p := recover(); p != nil {
			slog.Error(errs.WrapRecoveredPanic(p, debug.Stack()).Error())
			exitCode = 1
		}
		slog.Info("Shutdown complete", "exitCode", exitCode)
		os.Exit(exitCode)
	}()

	if err := command(); err != nil {
		slog.Error(errs.Wrap("command returned error", err).Error())
		slog.Warn("Exited abnormally")
		exitCode = 1
	}
}
