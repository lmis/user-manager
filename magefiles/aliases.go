package main

var Aliases = map[string]interface{}{
	"sta": Start,
	"b":   Build.App,
	"c":   Check,
	"fmt": Format,
	"cu":  ComposeUpLocalEnvironment,
	"cd":  ComposeDownLocalEnvironment,
	//"ft":  FunctionalTests.Basic,
}
