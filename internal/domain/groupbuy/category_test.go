package groupbuy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCategory_Valid(t *testing.T) {
	c, err := NewCategory("Clothing", []string{"Size", "Color"})
	require.NoError(t, err)
	assert.NotEmpty(t, c.ID)
	assert.Equal(t, "Clothing", c.Name)
	assert.Equal(t, []string{"Size", "Color"}, c.SpecNames)
}

func TestNewCategory_EmptyName(t *testing.T) {
	c, err := NewCategory("", []string{"Size"})
	assert.Error(t, err)
	assert.Nil(t, c)
}

func TestNewCategory_NilSpecs(t *testing.T) {
	c, err := NewCategory("Food", nil)
	require.NoError(t, err)
	assert.NotEmpty(t, c.ID)
	assert.Equal(t, "Food", c.Name)
	assert.Nil(t, c.SpecNames)
}
