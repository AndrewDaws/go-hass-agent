// Copyright (c) 2024 Joshua Rich <joshua.rich@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

type Package mg.Namespace

var (
	pkgformats   = []string{"rpm", "deb", "archlinux"}
	pkgPath      = filepath.Join(distPath, "pkg")
	nfpmBaseArgs = []string{"run", "github.com/goreleaser/nfpm/v2/cmd/nfpm", "package", "--config", ".nfpm.yaml", "--target", pkgPath}

	ErrNoBuildEnv        = errors.New("no build and/or version environment variables")
	ErrNfpmInstallFailed = errors.New("unable to install nfpm")
)

// Nfpm builds packages using nfpm.
//
//nolint:mnd,cyclop
func (Package) Nfpm() error {
	if err := os.RemoveAll(pkgPath); err != nil {
		return fmt.Errorf("could not clean dist directory: %w", err)
	}

	if err := os.MkdirAll(pkgPath, 0o755); err != nil {
		return fmt.Errorf("could not create dist directory: %w", err)
	}

	envMap, err := generateEnv()
	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}

	for _, pkgformat := range pkgformats {
		slog.Info("Building package with nfpm.", slog.String("format", pkgformat))
		args := slices.Concat(nfpmBaseArgs, []string{"--packager", pkgformat})

		if err := sh.RunWithV(envMap, "go", args...); err != nil {
			return fmt.Errorf("could not run nfpm: %w", err)
		}

		// nfpm creates the same package name for armv6 and armv7 deb packages,
		// so we need to rename them.
		if envMap["GOARCH"] == "arm" && pkgformat == "deb" {
			debPkgs, err := filepath.Glob(distPath + "/pkg/*.deb")
			if err != nil || debPkgs == nil {
				return fmt.Errorf("could not find arm deb package: %w", err)
			}

			oldDebPkg := debPkgs[0]
			newDebPkg := strings.ReplaceAll(oldDebPkg, "armhf", "arm"+envMap["GOARM"]+"hf")

			if err = sh.Copy(newDebPkg, oldDebPkg); err != nil {
				return fmt.Errorf("could not rename old arm deb package: %w", err)
			}

			err = sh.Rm(oldDebPkg)
			if err != nil {
				return fmt.Errorf("could not remove old arm deb package: %w", err)
			}
		}
	}

	return nil
}

// CI builds all packages as part of the CI pipeline.
func (p Package) CI() error {
	if !isCI() {
		return ErrNotCI
	}

	mg.Deps(p.Nfpm)

	return nil
}
