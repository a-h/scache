package data

import (
	"reflect"
	"testing"
)

func TestDataIDString(t *testing.T) {
	tests := []struct {
		name     string
		input    ID
		expected string
	}{
		{
			name:     "basic ID",
			input:    NewID("db.table.column", "12345"),
			expected: "id=12345&s=db.table.column&t=data.ID",
		},
		{
			name:     `ID which contains "\"`,
			input:    NewID(`\\`, "12345"),
			expected: `id=12345&s=%5C%5C&t=data.ID`,
		},
		{
			name:     `ID which contains "="`,
			input:    NewID(`db=test`, "12345"),
			expected: `id=12345&s=db%3Dtest&t=data.ID`,
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

func TestDataIDParseErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expected      ID
		expectedError error
	}{
		{
			name:          "No valid values",
			input:         "?something=123&else=12",
			expectedError: ErrNotDataID,
		},
		{
			name:          "Invalid URL escape",
			input:         "&%2+3=4&",
			expectedError: ErrMalformed,
		},
	}

	for _, test := range tests {
		id, err := Parse(test.input)
		if !reflect.DeepEqual(test.expectedError, err) {
			t.Fatalf("%s: expected error %v, got '%v'", test.name, test.expectedError, err)
		}
		if test.expected != id {
			t.Errorf("%s: Parse(): expected '%v', got '%v'", test.name, test.expected, id)
		}
	}
}
