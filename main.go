// Package main is the command-line tool to manage and optimize defacto2.net.
package main

/*
Copyright © 2021-23 Ben Garrett <code.by.ben@gmail.com>

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
	"runtime/debug"
	"strings"

	"github.com/Defacto2/df2/cmd"
	"github.com/Defacto2/df2/pkg/conf"
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
	// Logger, uses a preset config until the env vars are read
	l, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v\n", err)
	}
	defer func() {
		// see issue on false-positive errors.
		// https://github.com/uber-go/zap/issues/370
		_ = l.Sync()
	}()
	ls := l.Sugar()
	// Panic recovery to close any active connections and to log the problem
	defer func() {
		if i := recover(); i != nil {
			debug.PrintStack() // uncomment to trace
			ls.DPanic(i)
		}
	}()
	// Environment configuration
	configs := conf.Defaults()
	if err := env.Parse(
		&configs, conf.Options()); err != nil {
		ls.Fatalf("Environment variable probably contains an invalid value: %s.", err)
	}
	// Go runtime customizations
	setProcs(configs)
	// Setup the production logger
	if !configs.IsProduction {
		ls = logger.Production().Sugar()
	}
	// Configuration sanity checks
	if err := configs.Checks(ls); err != nil {
		ls.Error(err)
	}
	if ascii() {
		color.Enable = false
	}
	// Execute help and exit
	if help() {
		execHelp(ls, configs)
		return
	}
	// Print the compile and version details
	if progInfo() {
		execInfo(ls, configs)
		return
	}
	// Suppress stdout
	if quiet() {
		os.Stdout, _ = os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		defer os.Stdout.Close()
	}
	// Database check
	checkDB(ls, configs)
	// Execute the cobra flag library
	if err := cmd.Execute(ls, configs); err != nil {
		ls.Error(err)
		defer os.Exit(1)
	}
}

// setProcs sets the maximum number of CPUs that can be executing simultaneously.
func setProcs(c conf.Config) {
	if i := c.MaxProcs; i > 0 {
		runtime.GOMAXPROCS(int(i))
	}
}

// checkDB checks the database connection.
func checkDB(ls *zap.SugaredLogger, c conf.Config) {
	db, err := msql.Connect(c)
	if err != nil {
		ls.Errorf("Could not connect to the database: %s.", err)
	}
	defer func() {
		if db == nil {
			return
		}
		if !c.IsProduction {
			ls.Info("Closed the tcp connection to the database.")
		}
		if err := db.Close(); err != nil {
			ls.Error(err)
		}
	}()
}

// execInfo prints the compile and version details.
func execInfo(ls *zap.SugaredLogger, c conf.Config) {
	w := os.Stdout
	err := cmd.Brand(w, ls, brand)
	if err != nil {
		ls.Error(err)
	}
	s, err := cmd.ProgInfo(ls, c, version)
	if err != nil {
		ls.Error(err)
		return
	}
	fmt.Fprint(w, s)
}

// execHelp prints the help and exits.
func execHelp(ls *zap.SugaredLogger, c conf.Config) {
	if err := cmd.Execute(ls, c); err != nil {
		ls.Error(err)
		// use defer to close any open connections
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
