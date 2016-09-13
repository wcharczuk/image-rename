package lib

import "time"

// NewDateIndexCollector returns a new date index collector.
func NewDateIndexCollector() *DateIndexCollector {
	return &DateIndexCollector{
		ByYear:  map[int]int{},
		ByMonth: map[int]map[time.Month]int{},
		ByDay:   map[int]map[time.Month]map[int]int{},
	}
}

// DateIndexCollector returns indexes by various components of a date.
type DateIndexCollector struct {
	Count   int
	ByYear  map[int]int
	ByMonth map[int]map[time.Month]int
	ByDay   map[int]map[time.Month]map[int]int
}

// Len returns the total number of elements counted.
func (dtic *DateIndexCollector) Len() int {
	return dtic.Count
}

// Add increments relevant buckets for a timestamp.
func (dtic *DateIndexCollector) Add(timestamp time.Time) {
	dtic.Count++
	dtic.ByYear[timestamp.Year()]++

	if _, hasYear := dtic.ByMonth[timestamp.Year()]; !hasYear {
		dtic.ByMonth[timestamp.Year()] = map[time.Month]int{}
	}
	dtic.ByMonth[timestamp.Year()][timestamp.Month()]++

	if _, hasYear := dtic.ByDay[timestamp.Year()]; !hasYear {
		dtic.ByDay[timestamp.Year()] = map[time.Month]map[int]int{}
	}
	if _, hasMonth := dtic.ByDay[timestamp.Year()][timestamp.Month()]; !hasMonth {
		dtic.ByDay[timestamp.Year()][timestamp.Month()] = map[int]int{}
	}
	dtic.ByDay[timestamp.Year()][timestamp.Month()][timestamp.Day()]++
}

// GetIndexByYear returns the index by the year.
func (dtic DateIndexCollector) GetIndexByYear(timestamp time.Time) int {
	return dtic.ByYear[timestamp.Year()]
}

// GetIndexByMonth returns the index by the month.
func (dtic DateIndexCollector) GetIndexByMonth(timestamp time.Time) int {
	if months, hasYear := dtic.ByMonth[timestamp.Year()]; hasYear {
		if monthIndex, hasMonth := months[timestamp.Month()]; hasMonth {
			return monthIndex
		}
	}
	return 0
}

// GetIndexByDay returns the index by the day.
func (dtic DateIndexCollector) GetIndexByDay(timestamp time.Time) int {
	if months, hasYear := dtic.ByDay[timestamp.Year()]; hasYear {
		if days, hasMonth := months[timestamp.Month()]; hasMonth {
			if dayIndex, hasDay := days[timestamp.Day()]; hasDay {
				return dayIndex
			}
		}
	}
	return 0
}
