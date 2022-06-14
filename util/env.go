package util

import "os"

func GetEnvOrDefault(log Logger, envVar string, defaultVal string) string {
	res := os.Getenv(envVar)
	if res == "" {
		res = defaultVal
		log.Warn("Env variable \"%s\" not set, defaulting to %s", envVar, defaultVal)
	}
	return res
}
