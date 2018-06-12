package data

import "testing"

func TestDataStringRepresentation(t *testing.T) {
	tests := []struct {
		name     string
		input    ID
		expected string
	}{
		{
			name:     "basic ID",
			input:    NewID("db.table.column", "12345"),
			expected: "id=12345&source=db.table.column",
		},
		{
			name:     `ID which contains "\"`,
			input:    NewID(`\\`, "12345"),
			expected: `id=12345&source=%5C%5C`,
		},
		{
			name:     `ID which contains "="`,
			input:    NewID(`db=test`, "12345"),
			expected: `id=12345&source=db%3Dtest`,
		},
	}

	for _, test := range tests {
		as := test.input.String()
		if test.expected != as {
			t.Errorf("%s: String(): expected '%s', got '%s'", test.name, test.expected, as)
		}
		parsed, err := Parse(as)
		if err != nil {
			t.Fatalf("%s: Parse(): unexpected error: %v", test.name, err)
		}
		if parsed != test.input {
			t.Errorf("%s: Parse(): expected parsed value to equal input value", test.name)
		}
	}
}
