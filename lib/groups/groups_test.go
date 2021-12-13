package groups_test

import (
	"reflect"
	"testing"

	"github.com/Defacto2/df2/lib/groups"
)

func TestVariations(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name     string
		args     args
		wantVars []string
		wantErr  bool
	}{
		{"0", args{""}, []string{}, false},
		{"1", args{"hello"}, []string{"hello"}, false},
		{"2", args{"hello world"}, []string{
			"hello world",
			"helloworld", "hello-world", "hello_world", "hello.world",
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVars, err := groups.Variations(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Variations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotVars, tt.wantVars) {
				t.Errorf("Variations() = %v, want %v", gotVars, tt.wantVars)
			}
		})
	}
}

func TestFix(t *testing.T) {
	tests := []struct {
		name     string
		simulate bool
		wantErr  bool
	}{
		{"sim", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := groups.Fix(tt.simulate); (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
