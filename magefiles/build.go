package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Build mg.Namespace

// App checks and builds app
func (b Build) App() error {
	mg.Deps(Check)

	return sh.Run("go", "build", "-o", "bin/app", "cmd/app/main.go")
}

// EmailJob checks and builds email job
func (b Build) EmailJob() error {
	mg.Deps(Check)

	return sh.Run("go", "build", "-o", "bin/email-job", "cmd/email-job/main.go")
}

// MockApis checks and builds mock 3rd party apis
func (b Build) MockApis() error {
	mg.Deps(Check)

	return sh.Run("go", "build", "-o", "bin/mock-api", "cmd/mock-3rd-party-apis/main.go")
}
