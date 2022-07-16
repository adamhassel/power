package power

import (
	"testing"
	"time"

	"github.com/adamhassel/power/entities"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, 0, testdata.From.Minute())
	assert.Equal(t, 0, testdata.From.Second())
	assert.Equal(t, 0, testdata.To.Minute())
	assert.Equal(t, 0, testdata.To.Second())

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "simple, in window",
			fields: testdata,
			args: args{
				from: time.Now().Add(90 * time.Minute),
				to:   time.Now().Add(2 * time.Hour),
			},
			want: true,
		},
		{
			name:   "exactly the same as the boundaries",
			fields: testdata,
			args: args{
				from: testdata.From,
				to:   testdata.To,
			},
			want: true,
		},
		{
			name:   "both out of range",
			fields: testdata,
			args: args{
				from: time.Now().Add(24 * time.Hour),
				to:   time.Now().Add(25 * time.Hour),
			},
			want: false,
		},
		{
			name:   "from in, to out of range",
			fields: testdata,
			args: args{
				from: time.Now().Add(30 * time.Minute),
				to:   time.Now().Add(25 * time.Hour),
			},
			want: false,
		},
		{
			name:   "from out of, to in range",
			fields: testdata,
			args: args{
				from: time.Now().Add(-90 * time.Minute),
				to:   time.Now().Add(2 * time.Hour),
			},
			want: false,
		},
		{
			name:   "from is after to, but both are in the window",
			fields: testdata,
			args: args{
				to:   time.Now().Add(30 * time.Minute),
				from: time.Now().Add(1 * time.Hour),
			},
			want: false,
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

func TestFullPrices_Range(t *testing.T) {
	type args struct {
		from time.Time
		to   time.Time
	}

	now := time.Now().Truncate(time.Hour)

	testdata := FullPrices{
		Contents: []entities.FullPrice{
			{
				ValidFrom: now,
				ValidTo:   now.Add(time.Hour),
			},
			{
				ValidFrom: now.Add(time.Hour),
				ValidTo:   now.Add(2 * time.Hour),
			},
			{
				ValidFrom: now.Add(2 * time.Hour),
				ValidTo:   now.Add(3 * time.Hour),
			},
			{
				ValidFrom: now.Add(3 * time.Hour),
				ValidTo:   now.Add(4 * time.Hour),
			},
		},
		From: now,
		To:   now.Add(4 * time.Hour),
	}

	tests := []struct {
		name   string
		fields FullPrices
		args   args
		want   FullPrices
	}{
		{
			name:   "two middle hours",
			fields: testdata,
			args: args{
				from: now.Add(time.Hour),
				to:   now.Add(3 * time.Hour),
			},
			want: FullPrices{testdata.Contents[1:3], now.Add(time.Hour), now.Add(3 * time.Hour)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := FullPrices{
				Contents: tt.fields.Contents,
				From:     tt.fields.From,
				To:       tt.fields.To,
			}
			assert.Equalf(t, tt.want, fp.Range(tt.args.from, tt.args.to), "Range(%v, %v)", tt.args.from, tt.args.to)
		})
	}
}
