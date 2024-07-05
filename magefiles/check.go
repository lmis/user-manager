package main

import (
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
	"github.com/pkg/errors"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	assets               = "cmd/app/router/assets"
	buildDir             = "build"
	tailwindOut          = "cmd/app/router/assets/tailwind.css"
	templDir             = "cmd/app/router/render"
	templGeneratedSuffix = "_templ.go"

	templ        = "templ"
	golangciLint = "golangci-lint"
	wgo          = "wgo"
)

// Tidy runs go mod tidy
func Tidy() error {
	return sh.Run("go", "mod", "tidy")
}

// Tools installs all tools in tools.go
func Tools() error {
	mg.Deps(Tidy)

	if err := sh.Run("which", templ, golangciLint, wgo); err != nil {
		log.Print("Installing tools")
		buf, err := os.ReadFile("tools.go")
		if err != nil {
			return errors.Wrap(err, "error reading tools.go")
		}
		var tools []string
		for _, line := range strings.Split(string(buf), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "_") {
				tools = append(tools, strings.Trim(strings.TrimPrefix(trimmed, "_"), " \""))
			}
		}

		for _, tool := range tools {
			if err := sh.Run("go", "install", tool); err != nil {
				return errors.Wrapf(err, "error installing %s", tool)
			}
		}
	}
	return nil
}

// Templ generates templ files if dirty
func Templ() error {
	mg.Deps(Tidy, Tools)
	oldestDest := time.Time{}
	youngestSource := time.Time{}
	err := filepath.WalkDir(templDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		stat, err := os.Stat(path)
		if err != nil {
			return err
		}
		modTime := stat.ModTime()
		if strings.HasSuffix(path, templGeneratedSuffix) {
			if modTime.Before(oldestDest) || oldestDest.IsZero() {
				oldestDest = modTime
			}
		} else {
			if modTime.After(youngestSource) {
				youngestSource = modTime
			}

		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "error walking templ dir")
	}

	if youngestSource.After(oldestDest) {
		return sh.Run(templ, "generate")
	}
	log.Print("templ is up to date")
	return nil
}

// CleanTempl removes all generated templ files
func CleanTempl() error {
	return filepath.WalkDir(templDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, "_templ.go") {
			return os.Remove(path)
		}
		return nil
	})
}

// Format runs go fmt
func Format() error {
	mg.Deps(Tidy, Templ, Tailwind)

	return sh.Run("go", "fmt", "./...")
}

// Lint runs go linter
func Lint() error {
	mg.Deps(Tidy, Tools, Tailwind)

	return sh.Run(golangciLint, "run")
}

// Vet runs go vet
func Vet() error {
	mg.Deps(Tidy, Templ, Tailwind, Format)

	return sh.Run("go", "vet", "./...")
}

// BunInstall installs package.json dependencies
func BunInstall() error {
	return sh.Run("bun", "install")
}

// Tailwind generates tailwind.css if dirty
func Tailwind() error {
	mg.Deps(BunInstall)
	templates := "cmd/app/router/render"

	dirty, err := target.Dir(tailwindOut, templates, "main.css", "tailwind.config.js")
	if err != nil {
		return errors.Wrap(err, "error checking if tailwind.css is dirty")
	}
	if !dirty {
		log.Print("tailwind.css is up to date")
		return nil
	}
	// tailwindcss writes chatter to stderr, so we'll only show it if verbose
	var writer io.Writer
	if mg.Verbose() {
		writer = os.Stdout
	}
	_, err = sh.Exec(nil, writer, writer, "bun", "tailwindcss", "-i", "main.css", "-o", tailwindOut, "-m")
	if err != nil {
		return errors.Wrap(err, "error running tailwindcss")
	}
	// touch output file to make sure the timestamp is updated even if the content is the same
	return sh.Run("touch", tailwindOut)
}

// CleanBuildDir removes everything in build dir
func CleanBuildDir() error {
	return sh.Rm(path.Join(buildDir, "*"))
}

// CleanAssets removes everything in assets
func CleanAssets() error {
	return sh.Rm(path.Join(assets, "*"))
}

// BuildWebComponents builds web-components-all.js
func BuildWebComponents() error {
	mg.Deps(BunInstall)

	if err := sh.Run("mkdir", "-p", assets, buildDir); err != nil {
		return errors.Wrap(err, "error creating build dir and assets dir")
	}

	webComponentsInputFile := path.Join(buildDir, "web-components-all.ts")

	if err := sh.Run("sh", "-c", "cat cmd/app/router/render/web-components/*.ts > "+webComponentsInputFile); err != nil {
		return errors.Wrap(err, "error concatenating web-components")
	}
	if err := sh.Run("bun", "build", webComponentsInputFile, "--outdir", assets, "--minify"); err != nil {
		return errors.Wrap(err, "error building web-components")
	}
	return nil
}

// Check runs formats and checks
func Check() {
	mg.Deps(Tidy, Templ, Format, Lint, Vet, Tailwind, BuildWebComponents)
}

// Clean removes all generated files
func Clean() {
	mg.Deps(CleanAssets, CleanTempl, CleanBuildDir)
}
