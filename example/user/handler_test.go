package user

import (
	"net/http"
	"testing"
)

func TestGettingUserIDFromRequest(t *testing.T) {
	tests := []struct {
		method     string
		url        string
		expectedID string
	}{
		{
			method:     "GET",
			url:        "/user/123",
			expectedID: "123",
		},
		{
			method:     "GET",
			url:        "dev/user/123",
			expectedID: "123",
		},
		{
			method:     "GET",
			url:        "dev/user/123/",
			expectedID: "123",
		},
	}

	for i, test := range tests {
		r, err := http.NewRequest(test.method, test.url, nil)
		if err != nil {
			t.Errorf("test %d: unexpected error: %v", i, err)
			continue
		}
		actualID := getUserIDFromRequest(r)
		if actualID != test.expectedID {
			t.Errorf("test %d: expected ID '%s', got '%s'", i, test.expectedID, actualID)
		}
	}
}
