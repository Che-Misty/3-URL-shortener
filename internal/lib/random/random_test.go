package random

import (
	"testing"
)

func TestNewRandomString(t *testing.T) {
	tests := []struct {
		name string
		len  int
	}{
		{
			name: "len = 1",
			len:  1,
		},
		{
			name: "len = 5",
			len:  5,
		},
		{
			name: "len = 3",
			len:  3,
		},
		{
			name: "len = 10",
			len:  10,
		},
		{
			name: "len = 15",
			len:  15,
		},
		{
			name: "len = 17",
			len:  17,
		},
		{
			name: "len = 20",
			len:  20,
		},
		{
			name: "len = 25",
			len:  25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {})
	}
}
