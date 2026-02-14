package project

import (
	"errors"

	"github.com/google/uuid"
)

type Category struct {
	ID        string
	Name      string
	SpecNames []string
}

func NewCategory(name string, specNames []string) (*Category, error) {
	if name == "" {
		return nil, errors.New("category name is required")
	}
	return &Category{
		ID:        uuid.New().String(),
		Name:      name,
		SpecNames: specNames,
	}, nil
}
