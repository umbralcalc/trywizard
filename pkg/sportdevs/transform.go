package sportdevs

import (
	"strconv"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

func DataFrameFromMatchIncidents(mi *MatchesIncidentsResponse) dataframe.DataFrame {
	var times, types, classes []string
	var isHomes, awayScores, homeScores, revTimes, addedTimes []int
	var isLives []bool

	for _, e := range mi.Incidents {
		times = append(times, strconv.Itoa(e.Time))
		types = append(types, e.Type)
		classes = append(classes, e.Class)
		isHomes = append(isHomes, boolPtrToInt(e.IsHome))
		awayScores = append(awayScores, e.AwayScore)
		homeScores = append(homeScores, e.HomeScore)
		revTimes = append(revTimes, ptrToInt(e.ReversedPeriodTime))
		addedTimes = append(addedTimes, ptrToInt(e.AddedTime))
		isLives = append(isLives, ptrToBool(e.IsLive))
	}

	return dataframe.New(
		series.New(times, series.String, "time"),
		series.New(types, series.String, "type"),
		series.New(classes, series.String, "class"),
		series.New(isHomes, series.Int, "is_home"),
		series.New(awayScores, series.Int, "away_score"),
		series.New(homeScores, series.Int, "home_score"),
		series.New(revTimes, series.Int, "reversed_period_time"),
		series.New(addedTimes, series.Int, "added_time"),
		series.New(isLives, series.Bool, "is_live"),
	)
}

func ptrToInt(x *int) int {
	if x != nil {
		return *x
	}
	return 0
}
func ptrToBool(x *bool) bool {
	if x != nil {
		return *x
	}
	return false
}
func boolPtrToInt(x *bool) int {
	if x != nil && *x {
		return 1
	}
	return 0
}
