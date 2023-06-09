package utils

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"
)

func TestToken(t *testing.T) {
	testCase := []struct {
		userID string
	}{
		{userID: "userid-001"},
		{userID: "userid-002"},
	}

	for _, c := range testCase {
		token, err := GenerateToken(c.userID)
		assert.NilError(t, err)
		fmt.Println(token)
		resp, err := ParseToken(token)
		assert.NilError(t, err)
		fmt.Println(resp)
		assert.Equal(t, resp, c.userID)
		fmt.Println("------")
	}
}
