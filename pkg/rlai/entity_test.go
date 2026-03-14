package rlai

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithinRange(t *testing.T) {
	tests := []struct {
		name                   string
		x, y, z                int
		x2, y2, z2             int
		rangeX, rangeY, rangeZ int
		want                   bool
	}{
		{
			name: "exact match",
			x:    5, y: 5, z: 5,
			x2: 5, y2: 5, z2: 5,
			rangeX: 0, rangeY: 0, rangeZ: 0,
			want: true,
		},
		{
			name: "within positive range",
			x:    6, y: 7, z: 8,
			x2: 5, y2: 5, z2: 5,
			rangeX: 2, rangeY: 3, rangeZ: 4,
			want: true,
		},
		{
			name: "on negative edge of range",
			x:    3, y: 2, z: 1,
			x2: 5, y2: 5, z2: 5,
			rangeX: 2, rangeY: 3, rangeZ: 4,
			want: true,
		},
		{
			name: "outside x range",
			x:    8, y: 5, z: 5,
			x2: 5, y2: 5, z2: 5,
			rangeX: 2, rangeY: 0, rangeZ: 0,
			want: false,
		},
		{
			name: "outside y range",
			x:    5, y: 9, z: 5,
			x2: 5, y2: 5, z2: 5,
			rangeX: 0, rangeY: 3, rangeZ: 0,
			want: false,
		},
		{
			name: "outside z range",
			x:    5, y: 5, z: 10,
			x2: 5, y2: 5, z2: 5,
			rangeX: 0, rangeY: 0, rangeZ: 4,
			want: false,
		},
		{
			name: "on positive edge of range",
			x:    7, y: 8, z: 9,
			x2: 5, y2: 5, z2: 5,
			rangeX: 2, rangeY: 3, rangeZ: 4,
			want: true,
		},
		{
			name: "negative range (should not match)",
			x:    5, y: 5, z: 5,
			x2: 5, y2: 5, z2: 5,
			rangeX: -1, rangeY: -1, rangeZ: -1,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithinRange(tt.x, tt.y, tt.z, tt.x2, tt.y2, tt.z2, tt.rangeX, tt.rangeY, tt.rangeZ)
			assert.Equal(t, tt.want, got)
		})
	}
}
