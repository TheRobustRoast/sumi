package steps

import "testing"

func TestParseMigrationName(t *testing.T) {
	tests := []struct {
		filename string
		wantNum  string
		wantOk   bool
	}{
		{"0001-add-thing.sh", "0001", true},
		{"0023-fix-bug.sh", "0023", true},
		{"1-minimal.sh", "1", true},
		{"999-big-number.sh", "999", true},
		{"readme.txt", "", false},
		{"no-number.sh", "", false},
		{"-leading-dash.sh", "", false},
		{"abc-not-digits.sh", "", false},
		{"0001.sh", "", false},       // no dash separator
		{"0001-name.txt", "", false},  // wrong extension
		{"", "", false},
	}

	for _, tt := range tests {
		num, ok := ParseMigrationName(tt.filename)
		if num != tt.wantNum || ok != tt.wantOk {
			t.Errorf("ParseMigrationName(%q) = (%q, %v), want (%q, %v)",
				tt.filename, num, ok, tt.wantNum, tt.wantOk)
		}
	}
}
