package prods_test

import (
	"testing"

	"github.com/Defacto2/df2/lib/demozoo/internal/prods"
)

func TestProductionsAPIv1_Download(t *testing.T) {
	dl := prods.DownloadsAPIv1{
		LinkClass: "SceneOrgFile",
		URL:       "https://files.scene.org/view/parties/2000/ambience00/demo/feestje.zip",
	}
	tests := []struct {
		name    string
		p       prods.ProductionsAPIv1
		l       prods.DownloadsAPIv1
		wantErr bool
	}{
		{"empty", prods.ProductionsAPIv1{}, prods.DownloadsAPIv1{}, true},
		{"example1", example1, prods.DownloadsAPIv1{}, true},
		{"example1 dl", example1, dl, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Download(tt.l); (err != nil) != tt.wantErr {
				t.Errorf("ProductionsAPIv1.Download() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
