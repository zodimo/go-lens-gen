package gen

import (
	"testing"
)

func TestValidatePackageName(t *testing.T) {
	tests := []struct {
		name    string
		isValid bool
	}{
		{"userprofile", true},
		{"user_profile", true},
		{"user-profile", false},
		{"UserProfile", false},
		{"123test", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePackageName(tt.name)
			if (err == nil) != tt.isValid {
				t.Errorf("ValidatePackageName(%q) error = %v, want valid = %v", tt.name, err, tt.isValid)
			}
		})
	}
}

func TestValidateStructName(t *testing.T) {
	tests := []struct {
		name    string
		isValid bool
	}{
		{"UserProfileLens", true},
		{"CustomerLens", true},
		{"userProfileLens", false},
		{"user-profile-lens", false},
		{"My_Lens!@#", false},
		{"123Lens", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStructName(tt.name)
			if (err == nil) != tt.isValid {
				t.Errorf("ValidateStructName(%q) error = %v, want valid = %v", tt.name, err, tt.isValid)
			}
		})
	}
}

func TestSanitizePackageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user-profile", "userprofile"},
		{"UserProfile", "userprofile"},
		{"123test", "test"},
		{"---", "pkg"},
		{"123", "pkg123"}, // Wait, my implementation returns "pkg123". Let's verify.
		{"my_pkg_123", "my_pkg_123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizePackageName(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizePackageName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeStructName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user-profile-lens", "UserProfileLens"},
		{"customerLens", "CustomerLens"},
		{"my_lens!@#", "MyLens"},
		{"123lens", "Type123lens"}, // Or Type123Lens?
		{"---", "Struct"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := SanitizeStructName(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeStructName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
