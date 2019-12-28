/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

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
package cmd

import "testing"

func Test_options(t *testing.T) {
	type args struct {
		a []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{}, "\noptions: "},
		{"targets", args{targets}, "\noptions: all, download, emulation, image"},
		{"test", args{[]string{"test"}}, "\noptions: test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := options(tt.args.a); got != tt.want {
				t.Errorf("options() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_valid(t *testing.T) {
	type args struct {
		a []string
		x string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"empty", args{}, false},
		{"targets", args{targets, "all"}, true},
		{"no targets", args{targets, "foo"}, false},
		{"simple", args{[]string{"test"}, "test"}, true},
		{"empty x", args{[]string{"test"}, ""}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := valid(tt.args.a, tt.args.x); got != tt.want {
				t.Errorf("valid() = %v, want %v", got, tt.want)
			}
		})
	}
}
