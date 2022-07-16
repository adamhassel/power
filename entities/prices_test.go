package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFullPrice_InWindow(t *testing.T) {
	type fields struct {
		ValidFrom time.Time
		ValidTo   time.Time
	}
	type args struct {
		from time.Time
		to   time.Time
	}
	testdata := fields{
		ValidFrom: time.Now().Add(time.Hour).Truncate(time.Hour),
		ValidTo:   time.Now().Add(2 * time.Hour).Truncate(time.Hour),
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "Comfortably in window",
			fields: testdata,
			args: args{
				from: time.Now(),
				to:   time.Now().Add(4 * time.Hour),
			},
			want: true,
		},
		{
			name:   "Exactly in window",
			fields: testdata,
			args: args{
				from: testdata.ValidFrom,
				to:   testdata.ValidTo,
			},
			want: true,
		},
		{
			name:   "From is outside, to is inside",
			fields: testdata,
			args: args{
				from: time.Now().Add(90 * time.Minute),
				to:   time.Now().Add(2 * time.Hour),
			},
			want: false,
		},
		{
			name:   "From is inside, to is outside",
			fields: testdata,
			args: args{
				from: time.Now(),
				to:   time.Now().Add(1 * time.Hour),
			},
			want: false,
		},
		{
			name:   "both are outside",
			fields: testdata,
			args: args{
				from: time.Now().Add(61 * time.Minute),
				to:   time.Now().Add(62 * time.Minute),
			},
			want: false,
		},
		{
			name:   "to < from, but if swtiched, would be inside",
			fields: testdata,
			args: args{
				to:   time.Now(),
				from: time.Now().Add(4 * time.Hour),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := FullPrice{
				ValidFrom: tt.fields.ValidFrom,
				ValidTo:   tt.fields.ValidTo,
			}
			assert.Equalf(t, tt.want, fp.InWindow(tt.args.from, tt.args.to), "(%+v).InWindow(%v, %v)", tt.fields, tt.args.from, tt.args.to)
			//if got := fp.InWindow(tt.args.from, tt.args.to); got != tt.want {
			//				t.Errorf("InWindow() = %v, want %v", got, tt.want)
			//}
		})
	}
}
