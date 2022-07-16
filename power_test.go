package power

import (
	"testing"
	"time"

	"github.com/adamhassel/power/entities"
)

func TestFullPrices_InRange(t *testing.T) {
	type fields struct {
		Contents []entities.FullPrice
		From     time.Time
		To       time.Time
	}
	type args struct {
		from time.Time
		to   time.Time
	}

	testdata := fields{
		From: time.Now().Add(time.Hour).Truncate(time.Hour),
		To:   time.Now().Add(3 * time.Hour).Truncate(time.Hour),
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "simple",
			fields: testdata,
			args: args{
				from: time.Now(),
				to:   time.Now().Add(time.Hour),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb := FullPrices{
				Contents: tt.fields.Contents,
				From:     tt.fields.From,
				To:       tt.fields.To,
			}
			if got := fb.InRange(tt.args.from, tt.args.to); got != tt.want {
				t.Errorf("InRange() = %v, want %v", got, tt.want)
			}
		})
	}
}
