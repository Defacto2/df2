/*
Copyright © 2021 Ben Garrett <bengarrett77@gmail.com>

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
package main

import (
	"github.com/Defacto2/df2/lib/cmd"
	ver "github.com/Defacto2/df2/lib/version"
)

// goreleaser generated ldflags containers
// https://goreleaser.com/environment/#using-the-mainversion
var version, commit, date string

func main() {
	if version != "" {
		ver.B.Version = version
	}
	if commit != "" {
		ver.B.Commit = commit
	}
	if date != "" {
		ver.B.Date = date
	}
	cmd.Execute()
}
