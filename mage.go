//+build mage

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const binary = "./drunkenfall"
const dbFile = "./data/db.sql"

// psql runs a postgres command
var psql = sh.RunCmd("psql", "--host", "localhost", "--user", "postgres")

func version() string {
	version, err := sh.OutCmd("git", "describe", "--always", "--tags")()
	if err != nil {
		log.Fatal(err)
	}
	return version
}

func buildTime() string {
	return time.Now().Format(time.RFC3339)
}

type Build mg.Namespace

// Builds the backend service
func (Build) Drunkenfall() error {
	env := make(map[string]string)

	// These are needed for static linking to work inside Docker
	env["CGO_ENABLED"] = "0"
	env["GOOS"] = "linux"

	ldflags := fmt.Sprintf(
		"-w -X main.version=%s -X main.buildtime=%s",
		version(),
		buildTime(),
	)

	return sh.RunWith(
		env,
		"go", "build", "-v",
		"-ldflags", ldflags,
		"-o", binary,
	)
}

// Rebuilds node-sass.
//
// For whatever reason, `node-sass` has a compiled component that is intensely
// fragile to kernel updates, so it's alarmingly common that it breaks
// haphazardly and needs a rebuild. If you see `node-sass` runner errors,
// this is the thing you want to run.
func (Build) Sass() error {
	if err := os.Chdir("js"); err != nil {
		return err
	}

	defer os.Chdir("../")
	return sh.Run("npm", "rebuild", "node-sass")
}

type Run mg.Namespace

// Runs the backend service
func (Run) Drunkenfall() error {
	mg.Deps(Build.Drunkenfall)

	return sh.Run(binary)
}

// Runs the npm frontend
func (Run) Npm() error {
	if err := os.Chdir("js"); err != nil {
		return err
	}

	defer os.Chdir("../")
	return sh.Run("npm", "run", "dev")
}

// Runs the postgres service via docker-compose
func (Run) Postgres() error {
	return sh.Run("docker-compose", "up", "postgres")
}

// Runs the proxy, with localhost settings
func (Run) Proxy() error {
	mg.Deps(Proxy.Install)

	return sh.Run("sudo", "caddy", "-conf", "Caddyfile.local")
}

type Docker mg.Namespace

// Builds the backend service docker image
func (Docker) Drunkenfall() error {
	mg.Deps(Docker.Frontend)

	return sh.Run("docker-compose", "build", "drunkenfall")
}

// Builds the frontend docker image
func (Docker) Frontend() error {
	mg.Deps(Docker.SetFrontendVersion)

	err := sh.Run("docker-compose", "build", "frontend")
	if err != nil {
		return err
	}

	// Clean out previous build data
	err = sh.Run("rm", "-rf", "./dist/")
	if err != nil {
		return err
	}

	// Create a container and store the image ID
	imgid, err := sh.OutCmd("docker", "create", "drunkenfall_frontend:latest")()
	if err != nil {
		return err
	}

	// Move the files out from the container
	err = sh.Run("docker", "cp", fmt.Sprintf("%s:/dist", imgid), "./dist")
	if err != nil {
		return err
	}

	// Remove the container for cleanliness and then we're done
	return sh.Run("docker", "rm", imgid)
}

// Sets the current version in the frontend
func (Docker) SetFrontendVersion() error {
	data := fmt.Sprintf("export const version = '%s (%s)'\n", version(), buildTime())
	v, err := os.Open("js/src/version.js")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(v.Name(), []byte(data), 0644)
	if err != nil {
		return err
	}

	log.Printf("Frontend version set to \"%s\"", data)
	return nil
}

type Code mg.Namespace

// Run all backend tests
func (Code) Test() error {
	env := make(map[string]string)
	env["GIN_MODE"] = "test"

	return sh.RunWith(env, "go", "test", "-coverprofile=cover.out", "-v", "./towerfall")
}

// Install go dependencies
func (Code) Vendor() error {
	return sh.Run("dep", "ensure", "-v")
}

// Lints the backend code
func (Code) Lint() error {
	mg.Deps(Code.InstallLinter)

	return sh.Run("golangci-lint", "run", "--config", ".golangci-pedantic.yaml")
}

// Install the linter
func (Code) InstallLinter() error {
	path, err := exec.LookPath("golangci-lint")
	if err != nil {
		return err
	}

	if path != "" {
		log.Print("golangci-lint already installed - skipping installation")
		return nil
	}

	err = sh.Run("go", "get", "-coverprofile=cover.out", "-v", "./towerfall")
	if err != nil {
		return err
	}

	return nil
}

type Proxy mg.Namespace

// Downloads and installs the proxy
func (Proxy) Install() error {
	path, err := exec.LookPath("caddy")
	if err != nil {
		return err
	}

	if path != "" {
		log.Print("caddy already installed - skipping installation")
		return nil
	}

	err = sh.Run("go", "get", "github.com/mholt/caddy/caddy")
	if err != nil {
		return err
	}

	err = sh.Run("go", "get", "github.com/caddyserver/builds")
	if err != nil {
		return err
	}

	path, err = sh.OutCmd("go", "env", "GOPATH")()
	if err != nil {
		return err
	}

	if err := os.Chdir(fmt.Sprintf("%s/src/github.com/mholt/caddy/caddy", path)); err != nil {
		return err
	}

	return sh.Run("go", "run", "build.go")
}

type DB mg.Namespace

// Open a postgres shell
func (DB) Shell() error {
	return psql()
}

// Dumps the current database
func (DB) Dump() error {
	return sh.Run(
		"pg_dump",
		"--host", "localhost", "--user", "postgres",
		"--clean", "--create",
		dbFile,
	)
}

// Resets from a database dump
func (DB) Reset() error {
	err := psql("-c", "DROP DATABASE drunkenfall")
	if err != nil {
		return err
	}

	err = psql("-c", "CREATE DATABASE drunkenfall")
	if err != nil {
		return err
	}

	return nil
}
