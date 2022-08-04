package prom

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

func parseStartTimeAndEndTime(start, end string) (time.Time, time.Time, error) {
	var (
		minTime = time.Now().Add(-2 * time.Hour)
		maxTime = time.Now().Add(2 * time.Hour)
		err     error
	)

	stime := minTime
	etime := maxTime

	if start != "" {
		stime, err = parseTime(start)
		if err != nil {
			return stime, etime, fmt.Errorf("error parsing start time: %w", err)
		}
	}

	if end != "" {
		etime, err = parseTime(end)
		if err != nil {
			return stime, etime, fmt.Errorf("error parsing end time: %w", err)
		}
	}
	return stime, etime, nil
}

func parseTime(s string) (time.Time, error) {
	if t, err := strconv.ParseFloat(s, 64); err == nil {
		s, ns := math.Modf(t)
		return time.Unix(int64(s), int64(ns*float64(time.Second))).UTC(), nil
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("cannot parse %q to a valid timestamp", s)
}
