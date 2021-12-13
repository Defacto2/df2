package prod_test

import (
	"testing"
	"time"

	"github.com/Defacto2/df2/lib/demozoo/internal/prod"
)

func TestURL(t *testing.T) {
	type args struct {
		id int64
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"large", args{158411}, "https://demozoo.org/api/v1/productions/158411?format=json", false},
		{"small", args{1}, "https://demozoo.org/api/v1/productions/1?format=json", false},
		{"negative", args{-1}, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prod.URL(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("URL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("URL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProduction_URL(t *testing.T) {
	type fields struct {
		ID         int64
		Timeout    time.Duration
		link       string
		StatusCode int
		Status     string
	}
	tests := []struct {
		name     string
		fields   fields
		wantLink string
		wantErr  bool
	}{
		{"empty", fields{}, "https://demozoo.org/api/v1/productions/0?format=json", false},
		{"1", fields{ID: 1}, "https://demozoo.org/api/v1/productions/1?format=json", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &prod.Production{
				ID:         tt.fields.ID,
				Timeout:    tt.fields.Timeout,
				Link:       tt.fields.link,
				StatusCode: tt.fields.StatusCode,
				Status:     tt.fields.Status,
			}
			if err := p.URL(); (err != nil) != tt.wantErr {
				t.Errorf("Production.URL() error = %v, wantErr %v", err, tt.wantErr)
			}
			if p.Link != tt.wantLink {
				t.Errorf("Production.URL() link = %v, want %v", p.Link, tt.wantLink)
			}
		})
	}
}

func TestProduction_Get(t *testing.T) {
	type fields struct {
		ID         int64
		Timeout    time.Duration
		Link       string
		StatusCode int
		Status     string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{"empty", fields{}, "", false},
		{"invalid", fields{ID: -1}, "", true},
		{"okay", fields{ID: 1}, "Rob Is Jarig", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &prod.Production{
				ID:         tt.fields.ID,
				Timeout:    tt.fields.Timeout,
				Link:       tt.fields.Link,
				StatusCode: tt.fields.StatusCode,
				Status:     tt.fields.Status,
			}
			got, err := p.Get()
			if (err != nil) != tt.wantErr {
				t.Errorf("Production.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.Title != tt.want {
				t.Errorf("Production.Get() = %v, want %v", got.Title, tt.want)
			}
		})
	}
}
