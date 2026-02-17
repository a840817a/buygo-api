package user

// CheckIsManager is a helper to verify if a user is either the creator or a manager.
func CheckIsManager(creatorID string, managerIDs []string, userID string) bool {
	if creatorID == userID {
		return true
	}
	for _, id := range managerIDs {
		if id == userID {
			return true
		}
	}
	return false
}
