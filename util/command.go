package util

import (
	"os"
	"runtime/debug"
)

func Run(topic string, command func(Logger) error) {
	SetLogJSON(os.Getenv("LOG_JSON") != "")
	log := Log(topic)
	exitCode := 0

	defer func() {
		if p := recover(); p != nil {
			log.Recovery(p, debug.Stack())
			exitCode = 1
		}
		log.Info("Shutdown complete. ExitCode: %d", exitCode)
	}()

	if err := command(log); err != nil {
		log.Err(Wrap("main", "runMigrations returned error", err))
		log.Warn("Server exited abnormally")
		exitCode = 1
	}
}
