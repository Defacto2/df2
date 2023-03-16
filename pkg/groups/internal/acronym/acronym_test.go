package acronym_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/groups/internal/acronym"
)

func TestGet(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		{"", "", false},
		{"Defacto2", "DF2", false},
		{"not found", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := acronym.Get(nil, tt.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrim(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"", args{"Defacto (DF)"}, "Defacto"},
		{"", args{"Defacto 2 (DF2)"}, "Defacto 2"},
		{"", args{"Defacto 2"}, "Defacto 2"},
		{"", args{"Razor 1911 (RZR)"}, "Razor 1911"},
		{"", args{"Defacto (2) (DF2)"}, "Defacto (2)"},
		{"", args{"(Defacto 2) (DF2)"}, "(Defacto 2)"},
		{"", args{"Defacto(DF)"}, "Defacto(DF)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := acronym.Trim(tt.args.s); got != tt.want {
				t.Errorf("Trim() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGroup_Get(t *testing.T) {
	type fields struct {
		Name       string
		Initialism string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{"empty", fields{}, false},
		{"df2", fields{
			Name:       "defacto2",
			Initialism: "",
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &acronym.Group{
				Name:       tt.fields.Name,
				Initialism: tt.fields.Initialism,
			}
			if err := g.Get(nil); (err != nil) != tt.wantErr {
				t.Errorf("Group.Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
