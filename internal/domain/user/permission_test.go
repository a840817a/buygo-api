package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckIsManager(t *testing.T) {
	tests := []struct {
		name       string
		creatorID  string
		managerIDs []string
		userID     string
		expected   bool
	}{
		{
			name:       "user is creator",
			creatorID:  "user-1",
			managerIDs: []string{"user-2", "user-3"},
			userID:     "user-1",
			expected:   true,
		},
		{
			name:       "user is manager",
			creatorID:  "user-1",
			managerIDs: []string{"user-2", "user-3"},
			userID:     "user-2",
			expected:   true,
		},
		{
			name:       "user is another manager",
			creatorID:  "user-1",
			managerIDs: []string{"user-2", "user-3"},
			userID:     "user-3",
			expected:   true,
		},
		{
			name:       "user is neither creator nor manager",
			creatorID:  "user-1",
			managerIDs: []string{"user-2", "user-3"},
			userID:     "user-4",
			expected:   false,
		},
		{
			name:       "empty managers list and user is not creator",
			creatorID:  "user-1",
			managerIDs: []string{},
			userID:     "user-4",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckIsManager(tt.creatorID, tt.managerIDs, tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}
