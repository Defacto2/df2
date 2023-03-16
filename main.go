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
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/Defacto2/df2/cmd"
	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database/msql"
	"github.com/Defacto2/df2/pkg/logger"
	"github.com/caarlos0/env/v7"
	"github.com/gookit/color"
	"go.uber.org/zap"
)

//go:embed cmd/defacto2.txt
var brand []byte

//go:embed .version
var version string

func main() {
	// Logger (use a preset config until env are read)
	l, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer l.Sync()
	logr := l.Sugar()

	// Panic recovery to close any active connections and to log the problem.
	defer func() {
		if i := recover(); i != nil {
			//debug.PrintStack()
			logr.DPanic(i)
		}
	}()

	// Environment configuration
	configs := configger.Defaults()
	if err := env.Parse(
		&configs, configger.Options()); err != nil {
		logr.Fatalf("Environment variable probably contains an invalid value: %s.", err)
	}

	// Go runtime customizations
	if i := configs.MaxProcs; i > 0 {
		runtime.GOMAXPROCS(int(i))
	}

	// Setup the production logger
	if !configs.IsProduction {
		logr = logger.Production().Sugar()
	}

	// Configuration sanity checks
	configs.Checks(logr)
	if ascii() {
		color.Enable = false
	}

	// Execute help and exit
	if help() {
		if err := cmd.Execute(logr, configs); err != nil {
			logr.Error(err)
			// use defer to close any open connections
			defer os.Exit(1)
		}
		return
	}

	// Database check
	db, err := msql.Connect(configs)
	if err != nil {
		logr.Errorf("Could not connect to the mysql database: %s.", err)
	}
	defer func() {
		if !configs.IsProduction {
			logr.Info("Closed the tcp connection to the database.")
		}
		if err := db.Close(); err != nil {
			logr.Error(err)
		}
	}()

	// Print the compile and version details
	if progInfo() {
		w := os.Stdout
		cmd.Brand(w, logr, brand)
		s, err := cmd.ProgInfo(db, version)
		if err != nil {
			logr.Error(err)
			return
		}
		fmt.Fprint(w, s)
		return
	}

	// Suppress stdout
	if quiet() {
		os.Stdout, _ = os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		defer os.Stdout.Close()
	}

	// Execute the cobra flag library
	if err := cmd.Execute(logr, configs); err != nil {
		logr.Error(err)
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
