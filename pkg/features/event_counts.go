package features

import (
	"fmt"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

type MinuteRange struct {
	L int
	U int
}

var MatchMinuteSmoothingKernel = map[int]MinuteRange{
	0:  {L: 0, U: 5},
	1:  {L: 0, U: 5},
	2:  {L: 0, U: 5},
	3:  {L: 1, U: 6},
	4:  {L: 2, U: 7},
	5:  {L: 3, U: 8},
	6:  {L: 4, U: 9},
	7:  {L: 5, U: 10},
	8:  {L: 6, U: 11},
	9:  {L: 7, U: 12},
	10: {L: 8, U: 13},
	11: {L: 9, U: 14},
	12: {L: 10, U: 15},
	13: {L: 11, U: 16},
	14: {L: 12, U: 17},
	15: {L: 13, U: 18},
	16: {L: 14, U: 19},
	17: {L: 15, U: 20},
	18: {L: 16, U: 21},
	19: {L: 17, U: 22},
	20: {L: 18, U: 23},
	21: {L: 19, U: 24},
	22: {L: 20, U: 25},
	23: {L: 21, U: 26},
	24: {L: 22, U: 27},
	25: {L: 23, U: 28},
	26: {L: 24, U: 29},
	27: {L: 25, U: 30},
	28: {L: 26, U: 31},
	29: {L: 27, U: 32},
	30: {L: 28, U: 33},
	31: {L: 29, U: 34},
	32: {L: 30, U: 35},
	33: {L: 31, U: 36},
	34: {L: 32, U: 37},
	35: {L: 33, U: 38},
	36: {L: 34, U: 39},
	37: {L: 35, U: 40},
	38: {L: 35, U: 40},
	39: {L: 35, U: 40},
	40: {L: 35, U: 40},
	41: {L: 41, U: 46},
	42: {L: 41, U: 46},
	43: {L: 41, U: 46},
	44: {L: 42, U: 47},
	45: {L: 43, U: 48},
	46: {L: 44, U: 49},
	47: {L: 45, U: 50},
	48: {L: 46, U: 51},
	49: {L: 47, U: 52},
	50: {L: 48, U: 53},
	51: {L: 49, U: 54},
	52: {L: 50, U: 55},
	53: {L: 51, U: 56},
	54: {L: 52, U: 57},
	55: {L: 53, U: 58},
	56: {L: 54, U: 59},
	57: {L: 55, U: 60},
	58: {L: 56, U: 61},
	59: {L: 57, U: 62},
	60: {L: 58, U: 63},
	61: {L: 59, U: 64},
	62: {L: 60, U: 65},
	63: {L: 61, U: 66},
	64: {L: 62, U: 67},
	65: {L: 63, U: 68},
	66: {L: 64, U: 69},
	67: {L: 65, U: 70},
	68: {L: 66, U: 71},
	69: {L: 67, U: 72},
	70: {L: 68, U: 73},
	71: {L: 69, U: 74},
	72: {L: 70, U: 75},
	73: {L: 71, U: 76},
	74: {L: 72, U: 77},
	75: {L: 73, U: 78},
	76: {L: 74, U: 79},
	77: {L: 75, U: 80},
	78: {L: 76, U: 81},
	79: {L: 77, U: 82},
	80: {L: 78, U: 83},
	81: {L: 79, U: 84},
	82: {L: 79, U: 84},
	83: {L: 79, U: 84},
	84: {L: 79, U: 84},
}

func TransformToEventCounts(df *dataframe.DataFrame) dataframe.DataFrame {
	// Get unique minutes from original df
	uniqueMinutes := GetUniqueMinutes(df)

	// Get unique event types from original df
	uniqueEventTypes := GetUniqueEventTypes(df)

	// Grouping event counts by time
	gdf := df.GroupBy("minute", "event_type").Aggregation(
		[]dataframe.AggregationType{
			dataframe.Aggregation_COUNT,
		},
		[]string{"event_type"},
	)
	gdf = gdf.Arrange(dataframe.Sort("minute"))

	// Build a lookup from the grouped df
	lookup := make(map[string]float64) // key = time_int|event_type
	gTimeInts, _ := gdf.Col("minute").Int()
	gEventTypes := gdf.Col("event_type").Records()
	gCounts := gdf.Col("event_type_COUNT").Float()
	for i := range gTimeInts {
		key := fmt.Sprintf("%d|%s", gTimeInts[i], gEventTypes[i])
		lookup[key] = gCounts[i]
	}

	// Build new complete records with zero-filled gaps
	var (
		timeCol      []int
		eventTypeCol []string
		countCol     []float64
	)
	for _, t := range uniqueMinutes {
		for _, evt := range uniqueEventTypes {
			key := fmt.Sprintf("%d|%s", t, evt)
			count := 0.0
			if val, ok := lookup[key]; ok {
				count = val
			}
			timeCol = append(timeCol, t)
			eventTypeCol = append(eventTypeCol, evt)
			countCol = append(countCol, count)
		}
	}

	// Create densified DataFrame
	ddf := dataframe.New(
		series.New(eventTypeCol, series.String, "event_type"),
		series.New(countCol, series.Float, "event_type_COUNT"),
		series.New(timeCol, series.Int, "minute"),
	)
	return ddf
}

func TransformToSmoothedEventCounts(df *dataframe.DataFrame) dataframe.DataFrame {
	// Get the event counts DataFrame
	cdf := TransformToEventCounts(df)

	// Extract columns
	minutes, _ := cdf.Col("minute").Int()
	eventTypes := cdf.Col("event_type").Records()
	counts := cdf.Col("event_type_COUNT").Float()

	// Build map for lookup: key = minute|event_type, value = count
	countMap := make(map[string]float64)
	for i := range minutes {
		key := fmt.Sprintf("%d|%s", minutes[i], eventTypes[i])
		countMap[key] = counts[i]
	}

	// Build set of all unique minutes and event types
	minuteSet := make(map[int]struct{})
	eventTypeSet := make(map[string]struct{})
	for _, m := range minutes {
		minuteSet[m] = struct{}{}
	}
	for _, evt := range eventTypes {
		eventTypeSet[evt] = struct{}{}
	}

	// Convert to sorted slices
	var uniqueMinutes []int
	for m := range minuteSet {
		uniqueMinutes = append(uniqueMinutes, m)
	}
	var uniqueEventTypes []string
	for evt := range eventTypeSet {
		uniqueEventTypes = append(uniqueEventTypes, evt)
	}

	// Smoothed output slices
	var (
		smoothedMinutes    []int
		smoothedEventTypes []string
		smoothedCounts     []float64
	)

	// Perform smoothing
	for _, m := range uniqueMinutes {
		kernel, exists := MatchMinuteSmoothingKernel[m]
		if !exists {
			kernel = MinuteRange{L: m, U: m} // Default to no smoothing
		}
		for _, evt := range uniqueEventTypes {
			sum := 0.0
			count := 0.0
			for t := kernel.L; t <= kernel.U; t++ {
				key := fmt.Sprintf("%d|%s", t, evt)
				if val, ok := countMap[key]; ok {
					sum += val
					count++
				}
			}
			avg := 0.0
			if count > 0 {
				avg = sum / count
			}
			smoothedMinutes = append(smoothedMinutes, m)
			smoothedEventTypes = append(smoothedEventTypes, evt)
			smoothedCounts = append(smoothedCounts, avg)
		}
	}

	// Create and return smoothed DataFrame
	sdf := dataframe.New(
		series.New(smoothedEventTypes, series.String, "event_type"),
		series.New(smoothedCounts, series.Float, "smoothed_event_type_COUNT"),
		series.New(smoothedMinutes, series.Int, "minute"),
	)
	sdf = sdf.Arrange(dataframe.Sort("minute"))

	return sdf
}
