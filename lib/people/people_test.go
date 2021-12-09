package people

import (
	"reflect"
	"strings"
	"testing"

	"github.com/Defacto2/df2/lib/people/internal/role"
)

func TestFilters(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"", strings.Split(Roles(), ",")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filters(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name    string
		role    string
		want    int
		wantErr bool
	}{
		{"empty", "", 0, false},
		{"error", "error", 0, true},
		{"writers", "writers", 1, false},
		{"musicians", "m", 1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got, err := role.List(role.Roles(tt.role))
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error got = %v, want %v", (err != nil), tt.wantErr)
			}
			if got < tt.want {
				t.Errorf("List() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name    string
		r       Request
		wantErr bool
	}{
		{"empty", Request{}, false},
		{"unknown", Request{"unknown", false, true}, true},
		{"regular", Request{"writers", false, true}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Print(tt.r); (err != nil) != tt.wantErr {
				t.Errorf("Print() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func Test_DataList_HTML(t *testing.T) {
	type args struct {
		filename string
		r        Request
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"error", args{"", Request{"error", false, false}}, true},
		{"ok", args{"", Request{"", false, false}}, false},
		{"progress", args{"", Request{"", false, true}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := HTML(tt.args.filename, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("HTML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DataList(tt.args.filename, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("DataList() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRole_String(t *testing.T) {
	tests := []struct {
		name string
		r    role.Role
		want string
	}{
		{"err", -1, ""},
		{"all", 0, "all"},
		{"a", 1, "artists"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.String(); got != tt.want {
				t.Errorf("Role.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
