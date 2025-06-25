package features

import (
	"fmt"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

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
