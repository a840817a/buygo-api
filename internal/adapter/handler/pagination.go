package handler

import (
	"errors"
	"strconv"
)

func decodePageToken(token string) (int, error) {
	if token == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(token)
	if err != nil || offset < 0 {
		return 0, errors.New("invalid page token")
	}
	return offset, nil
}

func encodePageToken(offset int) string {
	if offset <= 0 {
		return ""
	}
	return strconv.Itoa(offset)
}

func normalizePageSize(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 100 {
		return 100
	}
	return limit
}
