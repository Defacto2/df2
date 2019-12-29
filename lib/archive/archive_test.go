package archive

import (
	"reflect"
	"testing"
)

func TestExtract(t *testing.T) {
	type args struct {
		name string
		uuid string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// In the future could create some dummy archive tests using []byte values
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Extract(tt.args.name, tt.args.uuid); (err != nil) != tt.wantErr {
				t.Errorf("Extract() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewExt(t *testing.T) {
	type args struct {
		name      string
		extension string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"none", args{"file", ""}, "file"},
		{"only ext", args{"", ".html"}, ".html"},
		{"txt", args{"file", ".txt"}, "file.txt"},
		{"spaces", args{"file name", ".docx"}, "file name.docx"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewExt(tt.args.name, tt.args.extension); got != tt.want {
				t.Errorf("NewExt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRead(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		// In the future could create some dummy archive tests using []byte values
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Read(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() = %v, want %v", got, tt.want)
			}
		})
	}
}
