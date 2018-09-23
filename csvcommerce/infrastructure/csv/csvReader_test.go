package csv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadCSV(t *testing.T) {
	rows, err := ReadProductCSV("fixture/nonexisting-products-file.csv")
	assert.Nil(t, rows)
	assert.Error(t, err)

	rows, err = ReadProductCSV("fixture/products.csv")
	assert.NoError(t, err, "no error expected")
	assert.NotNil(t, rows)
	assert.NotEmpty(t, rows)
}
