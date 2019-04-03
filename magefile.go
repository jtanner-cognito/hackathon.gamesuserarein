// +build mage

// This is the "magefile" for this service. For install instructions visit https://magefile.org/.
// To build the binaries for this service, just mage...
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	// mage:import go
	_ "github.com/CognitoIQ/cnp.shared.mage/golang"
	"github.com/magefile/mage/sh"

	// mage:import sam
	_ "github.com/CognitoIQ/cnp.shared.mage/sam"
	// mage:import precommit
	_ "github.com/CognitoIQ/cnp.shared.mage/precommit"
)

func init() {
	// We want Mage to always run with verbose
	os.Setenv("MAGEFILE_VERBOSE", "true")
}

// Builds the binaries in linux-amd64 [gb].
func Build() error {
	ps, err := listPackages()
	if err != nil {
		return err
	}

	for _, p := range ps {
		if err := sh.RunWith(flagEnv(), "go", "build", "-a", "-ldflags", "-s -w", "-o", fmt.Sprintf("%s/app", strings.Replace(p, "src", "bin", 1)), p); err != nil {
			return err
		}
		if err := sh.RunWith(flagEnv(), "$USERPROFILE/Go/bin/build-lambda-zip.exe", "-o", fmt.Sprintf("%s/app.zip", strings.Replace(p, "src", "bin", 1)), fmt.Sprintf("%s/app", strings.Replace(p, "src", "bin", 1))); err != nil {
			return err
		}
	}
	return nil
}

var (
	pkgs     []string
	pkgsInit sync.Once
)

func listPackages() ([]string, error) {
	pkgsInit.Do(func() {
		s, err := ioutil.ReadDir("./src/handlers")
		if err != nil {
			return
		}
		for _, p := range s {
			pkgs = append(pkgs, fmt.Sprintf("./src/handlers/%s", p.Name()))
		}
	})

	return pkgs, nil
}

var (
	libs     []string
	libsInit sync.Once
)

func listLibs() ([]string, error) {
	libsInit.Do(func() {
		s, err := ioutil.ReadDir("./lib")
		if err == nil {
			for _, p := range s {
				libs = append(libs, fmt.Sprintf("./lib/%s", p.Name()))
			}
		}
	})

	return libs, nil
}

func flagEnv() map[string]string {
	return map[string]string{
		"GOOS":       "linux",
		"GOARCH":     "amd64",
		"BUILD_DATE": time.Now().Format("2006-01-02T15:04:05Z0700"),
	}
}
