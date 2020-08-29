package csv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadCSV(t *testing.T) {
	rows, err := ReadCSV("fixture/nonexisting-products-file.csv")
	assert.Nil(t, rows)
	assert.Error(t, err)

	rows, err = ReadCSV("fixture/products2.csv")
	assert.NoError(t, err, "no error expected")
	assert.NotNil(t, rows)
	assert.NotEmpty(t, rows)
}
