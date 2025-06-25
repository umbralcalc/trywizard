package features

import (
	"os"
	"sort"
	"strconv"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

func GetRawEventsDataFrame() dataframe.DataFrame {
	file, _ := os.Open("../dat/events.csv")
	df := dataframe.ReadCSV(file)

	// Convert all times to ints
	timeStrs := df.Col("time").Records()
	timeInts := make([]int, len(timeStrs))
	for i, t := range timeStrs {
		val, err := strconv.Atoi(t[0 : len(t)-1])
		if err != nil {
			panic(err)
		}
		timeInts[i] = val
	}
	df = df.Mutate(series.New(series.Ints(timeInts), series.Int, "minute"))
	return df
}

func GetUniqueMinutes(df *dataframe.DataFrame) []int {
	// Get unique time_ints from original df
	timeIntsSet := make(map[int]struct{})
	for _, t := range df.Col("minute").Records() {
		val, _ := strconv.Atoi(t)
		timeIntsSet[val] = struct{}{}
	}
	uniqueTimeInts := make([]int, 0, len(timeIntsSet))
	for t := range timeIntsSet {
		uniqueTimeInts = append(uniqueTimeInts, t)
	}
	sort.Ints(uniqueTimeInts)
	return uniqueTimeInts
}

func GetUniqueEventTypes(df *dataframe.DataFrame) []string {
	// Get unique event types
	uniqueEventSet := make(map[string]struct{})
	for _, evt := range df.Col("event_type").Records() {
		uniqueEventSet[evt] = struct{}{}
	}
	uniqueEvents := make([]string, 0, len(uniqueEventSet))
	for evt := range uniqueEventSet {
		uniqueEvents = append(uniqueEvents, evt)
	}
	return uniqueEvents
}
