package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvent_IsManager_Creator(t *testing.T) {
	e := &Event{CreatorID: "creator-1"}
	assert.True(t, e.IsManager("creator-1"))
}

func TestEvent_IsManager_Manager(t *testing.T) {
	e := &Event{
		CreatorID:  "creator-1",
		ManagerIDs: []string{"mgr-1", "mgr-2"},
	}
	assert.True(t, e.IsManager("mgr-2"))
}

func TestEvent_IsManager_NoMatch(t *testing.T) {
	e := &Event{
		CreatorID:  "creator-1",
		ManagerIDs: []string{"mgr-1"},
	}
	assert.False(t, e.IsManager("random-user"))
}

func TestEvent_IsManager_EmptyManagers(t *testing.T) {
	e := &Event{CreatorID: "creator-1"}
	assert.False(t, e.IsManager("random-user"))
}
