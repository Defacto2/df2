package assets

import (
	"reflect"
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func Test_ignoreList(t *testing.T) {
	var want struct{}
	if got := ignoreList("")["blank.png"]; !reflect.DeepEqual(got, want) {
		t.Errorf("ignoreList() = %v, want %v", got, want)
	}

}
