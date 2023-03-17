package assets_test

import (
	"testing"

	"github.com/Defacto2/df2/pkg/assets"
	"github.com/Defacto2/df2/pkg/assets/internal/scan"
	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/directories"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gookit/color"
	"github.com/stretchr/testify/assert"
)

const empty = "empty"

func TestClean(t *testing.T) {
	type args struct {
		t      string
		delete bool
		human  bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"bad", args{"invalid", false, false}, true},
		{empty, args{}, true},
		{"good", args{"DOWNLOAD", false, false}, false},
	}
	cfg := configger.Defaults()
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := assets.Clean{
				Name:   tt.args.t,
				Remove: tt.args.delete,
				Human:  tt.args.human,
				Config: cfg,
			}
			if err := c.Walk(nil, nil); (err != nil) != tt.wantErr {
				t.Errorf("Clean() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateUUIDMap(t *testing.T) {
	tests := []struct {
		name      string
		wantTotal bool
		wantUuids bool
		wantErr   bool
	}{
		{"", true, true, false},
	}
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTotal, gotUuids, err := assets.CreateUUIDMap(nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUUIDMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (gotTotal > 0) != tt.wantTotal {
				t.Errorf("CreateUUIDMap() gotTotal = %v, want %v", gotTotal, tt.wantTotal)
			}
			if (len(gotUuids) > 0) != tt.wantUuids {
				t.Errorf("CreateUUIDMap() gotUuids = %v, want %v", len(gotUuids), tt.wantUuids)
			}
		})
	}
}

func TestCleaner(t *testing.T) {
	type args struct {
		t      assets.Target
		delete bool
		human  bool
	}
	cfg := configger.Defaults()
	d, err := directories.Init(cfg, false)
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"bad", args{-1, false, false}, true},
		{empty, args{}, false},
		{"good", args{assets.Download, false, false}, false},
	}
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := assets.Clean{
				Remove: tt.args.delete,
				Human:  tt.args.human,
				Config: cfg,
			}
			if err := c.Walker(nil, nil, tt.args.t, &d); (err != nil) != tt.wantErr {
				t.Errorf("Cleaner() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSkip(t *testing.T) {
	f, err := scan.Skip("", nil)
	assert.NotNil(t, err)
	assert.Equal(t, scan.Files{}, f)

	d, err := directories.Init(configger.Defaults(), false)
	assert.Nil(t, err)
	f, err = scan.Skip("", &d)
	assert.Nil(t, err)
	_, ok := f["blank.png"]
	assert.Equal(t, true, ok)
}

func TestTargets(t *testing.T) {
	const allTargets = 5
	tests := []struct {
		name   string
		target assets.Target
		want   int
	}{
		{"", assets.All, allTargets},
		{"", assets.Image, 2},
		{"error", -1, 0},
	}
	cfg := configger.Defaults()
	d, err := directories.Init(cfg, false)
	if err != nil {
		t.Error(err)
	}
	color.Enable = false
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := assets.Targets(cfg, tt.target, &d); len(got) != tt.want {
				t.Errorf("Targets() = %v, want %v", got, tt.want)
			}
		})
	}
}
