package prods_test

import (
	"reflect"
	"testing"

	"github.com/Defacto2/df2/pkg/demozoo/internal/prods"
)

func TestProductionsAPIv1_Authors(t *testing.T) {
	tests := []struct {
		name string
		p    prods.ProductionsAPIv1
		want prods.Authors
	}{
		{"empty", prods.ProductionsAPIv1{}, prods.Authors{}},
		{
			"record 1", example1,
			prods.Authors{nil, []string{"Ile"}, []string{"Ile"}, nil},
		},
		{
			"record 2", example2,
			prods.Authors{nil, []string{"Deep Freeze"}, []string{"The Cardinal"}, nil},
		},
		{
			"nick is_group", example3,
			prods.Authors{nil, nil, nil, nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Authors(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prods.ProductionsAPIv1.Authors() = %v, want %v", got, tt.want)
			}
		})
	}
}
