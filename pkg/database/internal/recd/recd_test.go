package recd_test

import (
	"bytes"
	"database/sql"
	"io"
	"os"
	"testing"
	"time"

	"github.com/Defacto2/df2/pkg/configger"
	"github.com/Defacto2/df2/pkg/database"
	"github.com/Defacto2/df2/pkg/database/internal/recd"
	"github.com/Defacto2/df2/pkg/directories"
	"github.com/Defacto2/df2/pkg/internal"
	"github.com/stretchr/testify/assert"
)

func TestRecord_String(t *testing.T) {
	r := recd.Record{}
	assert.NotEqual(t, "", r.String())
}

func TestRecord_Approve(t *testing.T) {
	r := recd.Record{}
	err := r.Approve(nil)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = r.Approve(db)
	assert.Nil(t, err)

	r = recd.Record{
		ID: 1,
	}
	err = r.Approve(db)
	assert.Nil(t, err)

	i := r.AutoID("1")
	assert.Equal(t, uint(1), i)
	err = r.Approve(db)
	assert.Nil(t, err)
}

func TestRecord_Check(t *testing.T) {
	r := recd.Record{}
	b, err := r.Check(nil, "", nil, nil)
	assert.NotNil(t, err)
	assert.False(t, b)

	dir, err := directories.Init(configger.Defaults(), false)
	assert.Nil(t, err)
	bb := &bytes.Buffer{}
	b, err = r.Check(bb, "", nil, &dir)
	assert.NotNil(t, err)
	assert.False(t, b)
	vals := make([]sql.RawBytes, 15)
	b, err = r.Check(bb, "", vals, &dir)
	assert.Nil(t, err)
	assert.False(t, b)
}

func TestRecord_Checks(t *testing.T) {
	r := recd.Record{}
	b := r.CheckDownload(io.Discard, "", "")
	assert.False(t, b)

	r = recd.Record{
		Filename: "test.zip",
	}
	r.CheckFileContent("")
	assert.False(t, b)

	b = r.CheckFileSize(internal.RandStr)
	assert.False(t, b)

	b = r.CheckFileSize("1024")
	assert.True(t, b)

	b = r.CheckImage(internal.RandStr)
	assert.False(t, b)
	b = r.CheckImage(internal.TestImg(4))
	assert.True(t, b)

	b = r.RecoverDownload(nil, "", "")
	assert.False(t, b)

	r = recd.Record{Filename: internal.Zip}
	f, err := os.CreateTemp(os.TempDir(), "recover-download")
	assert.Nil(t, err)
	defer f.Close()
	b = r.RecoverDownload(io.Discard, internal.TestArchives(4), f.Name())
	assert.True(t, b)
	defer os.Remove(f.Name())
}

func TestRecord_Summary(t *testing.T) {
	r := recd.Record{}
	r.Summary(nil, 0)
	r.Summary(nil, 1)
}

func TestNewApprove(t *testing.T) {
	b, err := recd.NewApprove(nil)
	assert.NotNil(t, err)
	assert.False(t, b)
	raw := make([]sql.RawBytes, 4)
	b, err = recd.NewApprove(raw)
	assert.Nil(t, err)
	assert.False(t, b)
	now := time.Now().Format(time.RFC3339)
	raw[2] = []byte(now)
	raw[3] = []byte(now)
	b, err = recd.NewApprove(raw)
	assert.Nil(t, err)
	assert.True(t, b)
}

func TestVerbose(t *testing.T) {
	recd.Verbose(io.Discard, false, "true")
	bb := &bytes.Buffer{}
	recd.Verbose(bb, true, internal.RandStr)
	assert.Contains(t, bb.String(), internal.RandStr)
}

func TestQueries(t *testing.T) {
	err := recd.Queries(nil, nil, configger.Config{}, false)
	assert.NotNil(t, err)

	db, err := database.Connect(configger.Defaults())
	assert.Nil(t, err)
	defer db.Close()
	err = recd.Queries(db, io.Discard, configger.Defaults(), false)
	assert.Nil(t, err)
}

func TestCheckGroups(t *testing.T) {
	type args struct {
		g1 string
		g2 string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"", args{"", ""}, false},
		{"", args{"CHANGEME", ""}, false},
		{"", args{"", "Changeme"}, false},
		{"", args{"A group", ""}, true},
		{"", args{"", "A group"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := recd.Record{}
			if got := r.CheckGroups(tt.args.g1, tt.args.g2); got != tt.want {
				t.Errorf("CheckGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValid(t *testing.T) {
	type args struct {
		deleted sql.RawBytes
		updated sql.RawBytes
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"new", args{[]byte("2006-01-02T15:04:05Z"), []byte("2006-01-02T15:04:05Z")}, true, false},
		{"new+offset", args{[]byte("2006-01-02T15:04:06Z"), []byte("2006-01-02T15:04:05Z")}, true, false},
		{"old del", args{[]byte("2016-01-02T15:04:05Z"), []byte("2006-01-02T15:04:05Z")}, false, false},
		{"old upd", args{[]byte("2000-01-02T15:04:05Z"), []byte("2016-01-02T15:04:05Z")}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := recd.Valid(tt.args.deleted, tt.args.updated)
			if (err != nil) != tt.wantErr {
				t.Errorf("Valid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Valid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReverseInt(t *testing.T) {
	tests := []struct {
		name         string
		value        uint
		wantReversed uint
	}{
		{"empty", 0, 0},
		{"count", 12345, 54321},
		{"seq", 555, 555},
		{"sign", 662211, 112266},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotReversed, _ := recd.ReverseInt(tt.value); gotReversed != tt.wantReversed {
				t.Errorf("ReverseInt() = %v, want %v", gotReversed, tt.wantReversed)
			}
		})
	}
}
