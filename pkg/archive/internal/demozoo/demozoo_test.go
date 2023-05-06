package demozoo_test

import (
	"os"
	"testing"

	"github.com/Defacto2/df2/pkg/archive/internal/content"
	"github.com/Defacto2/df2/pkg/archive/internal/demozoo"
	"github.com/Defacto2/df2/pkg/conf"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestData(t *testing.T) {
	t.Parallel()
	type fields struct {
		DOSee string
		NFO   string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"0", fields{"", ""}, "using \"\" for DOSee and \"\" as the NFO or text"},
		{"1a", fields{"hi.exe", ""}, "using \"hi.exe\" for DOSee and \"\" as the NFO or text"},
		{"1b", fields{"", "hi.txt"}, "using \"\" for DOSee and \"hi.txt\" as the NFO or text"},
		{"2", fields{"hi.exe", "hi.txt"}, "using \"hi.exe\" for DOSee and \"hi.txt\" as the NFO or text"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := demozoo.Data{
				DOSee: tt.fields.DOSee,
				NFO:   tt.fields.NFO,
			}
			if got := d.String(); got != tt.want {
				t.Errorf("Data.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTop(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		f    demozoo.Finds
		want string
	}{
		{"empty", demozoo.Finds{}, ""},
		{"1", demozoo.Finds{"file.exe": 0}, "file.exe"},
		{"2", demozoo.Finds{"file.exe": 0, "file.bat": 9}, "file.exe"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.f.Top(); got != tt.want {
				t.Errorf("Top() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDOS(t *testing.T) {
	t.Parallel()
	const (
		bat = ".bat"
		com = ".com"
		exe = ".exe"
	)
	f1 := content.File{
		Ext:        com,
		Name:       "hi.com",
		Executable: true,
	}
	f2 := content.File{
		Ext:        exe,
		Name:       "hi.exe",
		Executable: true,
	}
	f3 := content.File{
		Ext:        exe,
		Name:       "random.exe",
		Executable: true,
	}
	t1 := content.Contents{
		0: f1,
	}
	t2 := content.Contents{
		0: f1,
		1: f2,
	}
	t3 := content.Contents{
		0: f1,
		1: f2,
		2: f3,
	}
	t4 := content.Contents{
		0: f3,
	}
	vars := []string{"random"}
	none := []string{}

	f := demozoo.DOS(nil, "", nil, nil)
	assert.Equal(t, "", f)
	f = demozoo.DOS(nil, "hi.zip", t1, &vars)
	assert.Equal(t, "hi.com", f)
	f = demozoo.DOS(nil, "hi.zip", t2, &vars)
	assert.Equal(t, "hi.exe", f)
	f = demozoo.DOS(nil, "hi.zip", t3, &vars)
	assert.Equal(t, "hi.exe", f)
	f = demozoo.DOS(nil, "hi.zip", t4, &vars)
	assert.Equal(t, "random.exe", f)
	f = demozoo.DOS(nil, "hi.zip", t4, &none)
	assert.Equal(t, "random.exe", f)
}

func Test_MoveText(t *testing.T) {
	t.Parallel()
	err := demozoo.MoveText(nil, conf.Config{}, "", "")
	assert.ErrorIs(t, err, demozoo.ErrNoSrc)
	tmp := os.TempDir()
	err = demozoo.MoveText(nil, conf.Defaults(), tmp, "")
	assert.NotNil(t, err)
	err = demozoo.MoveText(nil, conf.Defaults(), tmp, database.TestID)
	assert.Nil(t, err)
}

func Test_NFO(t *testing.T) {
	t.Parallel()
	type args struct {
		name  string
		files content.Contents
	}
	var empty content.Contents
	var e []string
	f1 := content.File{
		Ext:      ".diz",
		Name:     "file_id.diz",
		Textfile: true,
	}
	f2 := content.File{
		Ext:      ".nfo",
		Name:     "hi.nfo",
		Textfile: true,
	}
	f3 := content.File{
		Ext:      ".txt",
		Name:     "random.txt",
		Textfile: true,
	}
	ff1 := make(content.Contents)
	ff1[0] = f1
	ff2 := make(content.Contents)
	ff2[0] = f1
	ff2[1] = f2
	ff3 := make(content.Contents)
	ff3[0] = f1
	ff3[1] = f2
	ff3[2] = f3
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", empty}, ""},
		{"empty zip", args{"hi.zip", empty}, ""},
		{"1 file", args{"hi.zip", ff1}, "file_id.diz"},
		{"2 file", args{"hi.zip", ff2}, "hi.nfo"},
		{"3 file", args{"hi.zip", ff2}, "hi.nfo"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := demozoo.NFO(tt.args.name, tt.args.files, &e); got != tt.want {
				t.Errorf("NFO() = %v, want %v", got, tt.want)
			}
		})
	}
}
