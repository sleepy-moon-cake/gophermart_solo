package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLuhnValid(t *testing.T) {
	tests := []struct {
		name   string
		number string
		want   bool
	}{
		{
			name:   "Valid number 1",
			number: "79927398713",
			want:   true,
		},
		{
			name:   "Valid number 2",
			number: "12345678903",
			want:   true,
		},
		{
			name:   "Invalid number (wrong checksum)",
			number: "79927398710",
			want:   false,
		},
		{
			name:   "Empty string",
			number: "",
			want:   false,
		},
		{
			name:   "String with letters",
			number: "7992739871A",
			want:   false,
		},
		{
			name:   "Very short invalid number",
			number: "1",
			want:   false,
		},
		{
			name:   "Short valid number",
			number: "0",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLuhnValid(tt.number)
			assert.Equal(t, tt.want, got, "IsLuhnValid(%s) failed", tt.number)
		})
	}
}
