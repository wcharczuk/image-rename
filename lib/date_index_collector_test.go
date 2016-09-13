package lib

import (
	"testing"
	"time"

	assert "github.com/blendlabs/go-assert"
)

func TestDateIndexCollector(t *testing.T) {
	assert := assert.New(t)

	collector := NewDateIndexCollector()
	collector.Add(time.Date(2015, 01, 01, 0, 0, 0, 0, time.UTC))
	collector.Add(time.Date(2016, 01, 01, 0, 0, 0, 0, time.UTC))
	collector.Add(time.Date(2016, 01, 02, 0, 0, 0, 0, time.UTC))
	collector.Add(time.Date(2016, 01, 02, 0, 0, 0, 0, time.UTC))
	collector.Add(time.Date(2016, 01, 03, 0, 0, 0, 0, time.UTC))
	collector.Add(time.Date(2016, 01, 03, 0, 0, 0, 0, time.UTC))
	collector.Add(time.Date(2016, 01, 03, 0, 0, 0, 0, time.UTC))
	collector.Add(time.Date(2016, 02, 01, 0, 0, 0, 0, time.UTC))
	assert.Equal(8, collector.Len())
	assert.Equal(7, collector.GetIndexByYear(time.Date(2016, 01, 01, 0, 0, 0, 0, time.UTC)))
	assert.Equal(6, collector.GetIndexByMonth(time.Date(2016, 01, 01, 0, 0, 0, 0, time.UTC)))
	assert.Equal(1, collector.GetIndexByDay(time.Date(2016, 01, 01, 0, 0, 0, 0, time.UTC)))
	assert.Equal(2, collector.GetIndexByDay(time.Date(2016, 01, 02, 0, 0, 0, 0, time.UTC)))
	assert.Equal(3, collector.GetIndexByDay(time.Date(2016, 01, 03, 0, 0, 0, 0, time.UTC)))
}
