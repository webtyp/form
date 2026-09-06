package form_test

import (
	"testing"

	"webtyp.com/input"
)

func TestIPValidation(t *testing.T) {
	ipInput := input.IP()

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		// Bug as reported
		{"uiuiuiu", "uiuiuiu", true},

		// Valid IPv4
		{"valid ipv4 - 1", "192.168.1.1", false},
		{"valid ipv4 - 2", "0.0.0.1", false},

		// Invalid IPv4
		{"invalid ipv4 out of range", "999.1.1.1", true},
		{"invalid ipv4 3 groups", "1.2.3", true},
		{"invalid ipv4 5 groups", "1.2.3.4.5", true},
		{"invalid ipv4 non-digit", "1.2.3.a", true},
		{"invalid ipv4 all zeros", "0.0.0.0", true},

		// Valid IPv6
		{"valid ipv6 short compression", "::1", false},
		{"valid ipv6 link local", "fe80::1", false},
		{"valid ipv6 long", "2001:db8::8a2e:370:7334", false},

		// Invalid IPv6
		{"invalid ipv6 non-hex", "gggg::1", true},
		{"invalid ipv6 too many groups", "2001:db8:1:2:3:4:5:6:7", true},

		// Mixed dot and colon
		{"mixed dot and colon", "::ffff:1.2.3.4", true},

		// Neither dot nor colon
		{"neither dot nor colon number", "12345", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ipInput.Validate(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr = %v", tt.value, err, tt.wantErr)
			}
		})
	}
}
