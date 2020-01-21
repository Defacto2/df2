package database

import (
	"testing"
)

func Test_record_imagePath(t *testing.T) {
	type fields struct {
		uuid string
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"empty", fields{"486070ae-f462-446f-b7e8-c70cb7a8a996"}, args{""}, "486070ae-f462-446f-b7e8-c70cb7a8a996.png"},
		{"ok", fields{"486070ae-f462-446f-b7e8-c70cb7a8a996"}, args{"/opt/temp/path"}, "/opt/temp/path/486070ae-f462-446f-b7e8-c70cb7a8a996.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := record{
				uuid: tt.fields.uuid,
			}
			if got := r.imagePath(tt.args.path); got != tt.want {
				t.Errorf("record.imagePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_record_checkGroups(t *testing.T) {
	type fields struct {
		c          int
		uuid       string
		filename   string
		filesize   uint
		zipContent string
		groupBy    string
		groupFor   string
		platform   string
		tag        string
		hashStrong string
		hashWeak   string
	}
	type args struct {
		g1 string
		g2 string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"", fields{}, args{"", ""}, false},
		{"", fields{}, args{"CHANGEME", ""}, false},
		{"", fields{}, args{"", "Changeme"}, false},
		{"", fields{}, args{"A group", ""}, true},
		{"", fields{}, args{"", "A group"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &record{
				c:          tt.fields.c,
				uuid:       tt.fields.uuid,
				filename:   tt.fields.filename,
				filesize:   tt.fields.filesize,
				zipContent: tt.fields.zipContent,
				groupBy:    tt.fields.groupBy,
				groupFor:   tt.fields.groupFor,
				platform:   tt.fields.platform,
				tag:        tt.fields.tag,
				hashStrong: tt.fields.hashStrong,
				hashWeak:   tt.fields.hashWeak,
			}
			if got := r.checkGroups(tt.args.g1, tt.args.g2); got != tt.want {
				t.Errorf("record.checkGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}
