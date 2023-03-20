package groups_test

import (
	"io"
	"reflect"
	"testing"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/groups"
	"github.com/stretchr/testify/assert"
)

func TestRequest_DataList(t *testing.T) {
	r := groups.Request{}
	err := r.DataList(nil, nil, nil)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = r.DataList(db, io.Discard, nil)
	assert.NotNil(t, err)
}

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
			gotVars, err := groups.Variations(nil, tt.args.name)
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
		name    string
		wantErr bool
	}{
		{"fix", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := groups.Fix(nil, nil); (err != nil) != tt.wantErr {
				t.Errorf("Fix() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
