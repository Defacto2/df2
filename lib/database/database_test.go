package database

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func Test_readPassword(t *testing.T) {

	// create a temporary file with the content EXAMPLE
	tmpFile, err := ioutil.TempFile(os.TempDir(), "prefix")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	text := []byte("EXAMPLE")
	if _, err = tmpFile.Write(text); err != nil {
		log.Fatal("Failed to write to temporary file", err)
	}
	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}

	tests := []struct {
		name   string
		pwPath string
		want   string
	}{
		{"empty", "", "password"},
		{"invalid", "/tmp/Test_readPassword/validfile", "password"},
		{"temp", tmpFile.Name(), "EXAMPLE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pwPath = tt.pwPath
			if got := readPassword(); got != tt.want {
				t.Errorf("readPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sqlGroupsWhere(t *testing.T) {
	type args struct {
		name           string
		incSoftDeletes bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"mag-", args{"magazine", false}, "AND section = 'magazine' AND `deletedat` IS NULL"},
		{"bbs-", args{"bbs", false}, "AND RIGHT(group_brand_for,4) = ' BBS' AND `deletedat` IS NULL"},
		{"ftp-", args{"ftp", false}, "AND RIGHT(group_brand_for,4) = ' FTP' AND `deletedat` IS NULL"},
		{"grp-", args{"group", false}, "AND RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine' AND `deletedat` IS NULL"},
		{"mag+", args{"magazine", true}, "AND section = 'magazine'"},
		{"bbs+", args{"bbs", true}, "AND RIGHT(group_brand_for,4) = ' BBS'"},
		{"ftp+", args{"ftp", true}, "AND RIGHT(group_brand_for,4) = ' FTP'"},
		{"grp+", args{"group", true}, "AND RIGHT(group_brand_for,4) != ' FTP' AND RIGHT(group_brand_for,4) != ' BBS' AND section != 'magazine'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sqlGroupsWhere(tt.args.name, tt.args.incSoftDeletes); got != tt.want {
				t.Errorf("sqlGroupsWhere() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_sqlGroups(t *testing.T) {
	type args struct {
		name               string
		includeSoftDeletes bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"all-", args{"all", false}, "(SELECT DISTINCT group_brand_for AS pubCombined FROM files WHERE Length(group_brand_for) <> 0 AND `deletedat` IS NULL) UNION (SELECT DISTINCT group_brand_by AS pubCombined FROM files WHERE Length(group_brand_by) <> 0 AND `deletedat` IS NULL) ORDER BY pubCombined"},
		{"all+", args{"all", true}, "(SELECT DISTINCT group_brand_for AS pubCombined FROM files WHERE Length(group_brand_for) <> 0 ) UNION (SELECT DISTINCT group_brand_by AS pubCombined FROM files WHERE Length(group_brand_by) <> 0 ) ORDER BY pubCombined"},
		{"ftp-", args{"ftp", false}, "SELECT DISTINCT group_brand_for AS pubCombined FROM files WHERE Length(group_brand_for) <> 0 AND RIGHT(group_brand_for,4) = ' FTP' AND `deletedat` IS NULL) ORDER BY pubCombined"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sqlGroups(tt.args.name, tt.args.includeSoftDeletes); got != tt.want {
				t.Errorf("sqlGroups() = %q, want %q", got, tt.want)
			}
		})
	}
}

func BenchmarkGroupsToHTML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GroupsToHTML("all", false, true, "")
	}
}
