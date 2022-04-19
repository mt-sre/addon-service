//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/magefile/mage/target"
)

const (
	protocGenGoVersion          = "1.28.0"
	protocGenGoGRPCVersion      = "1.2.0"
	protocGenGRPCGatewayVersion = "2.10.0"
	bufVersion                  = "1.3.1"
	goimportsVersion            = "0.1.5"
	golangCILintVersion         = "1.43.0"
)

var (
	workDir        string
	depsDir        string
	depsBinDir     string
	depsIncludeDir string
)

func init() {
	workDir, _ = os.Getwd()
	depsDir = path.Join(workDir, ".deps")
	depsBinDir = path.Join(depsDir, "bin")
	depsIncludeDir = path.Join(depsDir, "/include")
	os.Setenv("PATH", depsBinDir+":"+os.Getenv("PATH"))
}

func Generate() {
	mg.Deps(Build.Proto)
}

type Dependency mg.Namespace

func (Dependency) All() {
	mg.Deps(
		mg.F(Dependency.GoInstall, "buf", "github.com/bufbuild/buf/cmd/buf", bufVersion),
		mg.F(Dependency.GoInstall, "buf", "github.com/bufbuild/buf/cmd/protoc-gen-buf-breaking", bufVersion),
		mg.F(Dependency.GoInstall, "buf", "github.com/bufbuild/buf/cmd/protoc-gen-buf-lint", bufVersion),
		mg.F(Dependency.GoInstall, "protoc-gen-go", "google.golang.org/protobuf/cmd/protoc-gen-go", protocGenGoVersion),
		mg.F(Dependency.GoInstall, "protoc-gen-go-grpc", "google.golang.org/grpc/cmd/protoc-gen-go-grpc", protocGenGoGRPCVersion),
		mg.F(Dependency.GoInstall, "protoc-gen-grpc-gateway", "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway", protocGenGRPCGatewayVersion),
		mg.F(Dependency.GoInstall, "goimports", "golang.org/x/tools/cmd/goimports", goimportsVersion),
		mg.F(Dependency.GoInstall, "golangci-lint", "github.com/golangci/golangci-lint/cmd/golangci-lint", golangCILintVersion),
	)
}

func (Dependency) Dirs() error {
	// ensure deps dir
	if err := os.MkdirAll(".cache", os.ModePerm); err != nil {
		return fmt.Errorf("creating deps dir: %w", err)
	}
	if err := os.MkdirAll(".deps/bin", os.ModePerm); err != nil {
		return fmt.Errorf("creating deps/bin dir: %w", err)
	}
	return nil
}

func (Dependency) GoInstall(bin, packageURl, version string) error {
	mg.Deps(Dependency.Dirs)

	needsRebuild, err := checkBinDependencyNeedsRebuild(bin, version)
	if err != nil {
		return err
	}
	if !needsRebuild {
		return nil
	}

	url := packageURl + "@v" + version
	if err := sh.RunWithV(map[string]string{
		"GOBIN": depsBinDir,
	}, mg.GoCmd(),
		"install", url,
	); err != nil {
		return fmt.Errorf("install %s: %w", url, err)
	}
	return nil
}

// ensure a file and it's file path exist.
func ensureFile(file string) error {
	dir := filepath.Dir(file)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return fmt.Errorf("creating directory %s: %w", dir, err)
	}

	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		f, err := os.Create(file)
		if err != nil {
			return fmt.Errorf("creating file %s: %w", file, err)
		}
		defer f.Close()
		return nil
	}
	if err != nil {
		return fmt.Errorf("checking file %s: %w", file, err)
	}
	return nil
}

func checkBinDependencyNeedsRebuild(thing, version string) (needsRebuild bool, err error) {
	versionFile := ".deps/versions/v" + version
	if err := ensureFile(versionFile); err != nil {
		return false, fmt.Errorf("ensure file: %w", err)
	}

	rebuild, err := target.Path(".deps/bin/"+thing, versionFile)
	if err != nil {
		return false, fmt.Errorf("check: %w", err)
	}
	if !rebuild {
		return false, nil
	}

	return true, nil
}

type Build mg.Namespace

func (Build) Proto() error {
	mg.Deps(Dependency.All)

	err := sh.RunWithV(map[string]string{}, "buf", "generate", "--path", "api")
	if err != nil {
		return fmt.Errorf("running buf: %w", err)
	}
	return nil
}

func (Build) APIServer() {
	mg.SerialDeps(
		Dependency.All,
		mg.F(Build.Cmd, "apiserver"),
	)
}

func (Build) Cmd(cmd string) error {
	env := map[string]string{}
	if err := sh.RunWithV(env, "go", "build", "-v", "-o", "bin/"+cmd, "./cmd/"+cmd+"/main.go"); err != nil {
		return fmt.Errorf("compiling cmd/%s: %w", cmd, err)
	}
	return nil
}
