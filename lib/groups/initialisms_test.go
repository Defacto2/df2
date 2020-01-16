package groups

import (
	"testing"
)

// func Test_group_get(t *testing.T) {
// 	type fields struct {
// 		name       string
// 		initialism string
// 	}
// 	tests := []struct {
// 		name    string
// 		fields  fields
// 		want    string
// 		wantErr bool
// 	}{
// 		{"empty", fields{}, "", true},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			g := &group{
// 				name:       tt.fields.name,
// 				initialism: tt.fields.initialism,
// 			}
// 			if g.get(); g.initialism != tt.want {
// 				t.Errorf("group.initialism = %v, want %v", g.initialism, tt.want)
// 			}
// 			if err := g.get(); (err != nil) != tt.wantErr {
// 				t.Errorf("group.get() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 		})
// 	}
// }

func TestInitialism(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Initialism(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Initialism() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Initialism() = %v, want %v", got, tt.want)
			}
		})
	}
}
