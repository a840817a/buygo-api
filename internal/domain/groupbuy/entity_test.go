package groupbuy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGroupBuy_IsManager_Creator(t *testing.T) {
	gb := &GroupBuy{CreatorID: "creator-1"}
	assert.True(t, gb.IsManager("creator-1"))
}

func TestGroupBuy_IsManager_Manager(t *testing.T) {
	gb := &GroupBuy{
		CreatorID:  "creator-1",
		ManagerIDs: []string{"mgr-1", "mgr-2"},
	}
	assert.True(t, gb.IsManager("mgr-2"))
}

func TestGroupBuy_IsManager_NoMatch(t *testing.T) {
	gb := &GroupBuy{
		CreatorID:  "creator-1",
		ManagerIDs: []string{"mgr-1"},
	}
	assert.False(t, gb.IsManager("random-user"))
}

func TestGroupBuy_IsManager_EmptyManagers(t *testing.T) {
	gb := &GroupBuy{CreatorID: "creator-1"}
	assert.False(t, gb.IsManager("random-user"))
}
