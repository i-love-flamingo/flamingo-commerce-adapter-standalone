package csv

import (
	"testing"

	"fmt"

	"github.com/stretchr/testify/assert"
)

func TestReadCSV(t *testing.T) {
	rows, err := ReadProductCSV("fixture/nonexist-products.csv")
	assert.Nil(t, rows)
	assert.Error(t, err)

	rows, err = ReadProductCSV("fixture/products.csv")
	assert.NoError(t, err, "no error expected")
	assert.NotNil(t, rows)
	for _, row := range rows {
		for k, v := range row {
			fmt.Println("#########")
			fmt.Println(k + "+++++++++++>>>  " + v)
		}

	}

	assert.True(t, true)
}
