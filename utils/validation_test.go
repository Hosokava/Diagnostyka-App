package utils

import "testing"

func TestIsValidPESEL(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"Valid PESEL", "06301268369", true},
		{"Wrong checksum", "06301268360", false},
		{"Too short", "0630126", false},
		{"Too long", "063012683699", false},
		{"Empty", "", false},
		{"All zeros", "00000000000", true}, // Valid checksum
		{"Letters", "0630126836a", false},
		{"Spaces", "0630126 369", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsValidPESEL(tc.input)
			if got != tc.want {
				t.Fatalf("IsValidPESEL(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
