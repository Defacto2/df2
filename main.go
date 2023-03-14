package main

/*
Copyright © 2021-22 Ben Garrett <code.by.ben@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	_ "embed"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/Defacto2/df2/cmd"
	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database/msql"
	"github.com/Defacto2/df2/pkg/logger"
	"github.com/caarlos0/env/v7"
	"github.com/gookit/color"
)

//go:embed .version
var version string

func main() {
	// Logger (use the development log until the environment vars are parsed)
	log := logger.Development().Sugar()

	// Environment configuration
	configs := configger.Config{
		// hardcoded overrides can go here
		// IsProduction: true,
	}
	if err := env.Parse(
		&configs, env.Options{Prefix: configger.EnvPrefix}); err != nil {
		log.Fatalf("Environment variable probably contains an invalid value: %s.", err)
	}

	// Go runtime customizations
	if i := configs.MaxProcs; i > 0 {
		runtime.GOMAXPROCS(int(i))
	}

	// Setup the logger
	if configs.IsProduction {
		log = logger.Production().Sugar()
	}

	// Configuration sanity checks
	configs.Checks(log)
	if ascii() {
		color.Enable = false
	}

	// Database check
	db, err := msql.ConnectDB()
	if err != nil {
		log.Errorf("Could not connect to the mysql database: %s.", err)
	}
	defer db.Close()

	// print the compile and version details
	if progInfo() {
		s, err := cmd.ProgInfo(version)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprintln(os.Stdout, s)
		return
	}

	// suppress stdout except when requesting help
	if quiet() && !help() {
		os.Stdout, _ = os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		defer os.Stdout.Close()
	}

	// cobra flag library
	if err := cmd.Execute(); err != nil {
		log.Errorln(err)
		defer os.Exit(1)
	}
}

// global flags that should not be handled by the Cobra library
// to keep things simple, avoid using the flag standard library

// ascii returns true if the -ascii flag is in use.
func ascii() bool {
	for _, f := range os.Args {
		switch strings.ToLower(f) {
		case "-ascii", "--ascii":
			return true
		}
	}
	return false
}

// help returns true if the -help flag or alias is in use.
func help() bool {
	for _, f := range os.Args {
		switch strings.ToLower(f) {
		case "-h", "--h", "-help", "--help":
			return true
		}
	}
	return false
}

// progInfo returns true if the -version flag or alias is in use.
func progInfo() bool {
	for _, f := range os.Args {
		switch strings.ToLower(f) {
		case "-v", "--v", "-version", "--version":
			return true
		}
	}
	return false
}

// quiet returns true if the -quiet flag is in use.
func quiet() bool {
	for _, f := range os.Args {
		switch strings.ToLower(f) {
		case "-quiet", "--quiet":
			return true
		}
	}
	return false
}
