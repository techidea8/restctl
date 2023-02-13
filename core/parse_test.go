package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegPatern(t *testing.T) {
	times := map[string]string{
		"2023-03-04 13:15:02": "2006-01-02 15:04:05",
		"2023-03-04T13:15:02": "2006-01-02 15:04:05",
		"2023-03-04":          "2006-01-02",
	}
	for d, f := range times {
		timeformat := PredictTimeFormat(d)
		assert.Equal(t, f, timeformat)
	}
}
