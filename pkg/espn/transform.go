package espn

import (
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
)

func EventsDataFrameFromSummary(sum *Summary) dataframe.DataFrame {
	var times []string
	var eventTypes []string
	var eventTypeIDs []string
	var playerIDs []string
	var playerNames []string
	var teamIDs []string
	var homeScores []int
	var awayScores []int

	for _, comp := range sum.Header.Competitions {
		for _, event := range comp.Details {
			times = append(times, event.Clock.DisplayValue)
			eventTypes = append(eventTypes, event.Type.Text)
			eventTypeIDs = append(eventTypeIDs, event.Type.ID)

			// Get the first participant (usually the scorer/player)
			if len(event.Participants) > 0 && event.Participants[0].Athlete.ID != "" {
				playerIDs = append(playerIDs, event.Participants[0].Athlete.ID)
				playerNames = append(playerNames, event.Participants[0].Athlete.DisplayName)
			} else {
				playerIDs = append(playerIDs, "")
				playerNames = append(playerNames, "")
			}

			teamIDs = append(teamIDs, event.Team.ID)
			homeScores = append(homeScores, event.HomeScore)
			awayScores = append(awayScores, event.AwayScore)
		}
	}

	df := dataframe.New(
		series.New(times, series.String, "time"),
		series.New(eventTypes, series.String, "event_type"),
		series.New(eventTypeIDs, series.String, "event_type_id"),
		series.New(playerIDs, series.String, "player_id"),
		series.New(playerNames, series.String, "player_name"),
		series.New(teamIDs, series.String, "team_id"),
		series.New(homeScores, series.Int, "home_score"),
		series.New(awayScores, series.Int, "away_score"),
	)

	return df
}

func PlayersDataFrameFromSummary(sum *Summary) dataframe.DataFrame {
	var teamIDs []string
	var teamNames []string
	var playerIDs []string
	var playerNames []string
	var jerseyNumbers []string
	var positions []string
	var positionIDs []string
	var homeAways []string

	for _, rosterEntry := range sum.Rosters {
		team := rosterEntry.Team
		for _, player := range rosterEntry.Roster {
			teamIDs = append(teamIDs, team.ID)
			teamNames = append(teamNames, team.DisplayName)
			playerIDs = append(playerIDs, player.Athlete.ID)
			playerNames = append(playerNames, player.Athlete.DisplayName)
			jerseyNumbers = append(jerseyNumbers, player.Jersey)
			positions = append(positions, player.Position.DisplayName)
			positionIDs = append(positionIDs, player.Position.ID)
			homeAways = append(homeAways, rosterEntry.HomeAway)
		}
	}

	df := dataframe.New(
		series.New(teamIDs, series.String, "team_id"),
		series.New(teamNames, series.String, "team_name"),
		series.New(playerIDs, series.String, "player_id"),
		series.New(playerNames, series.String, "player_name"),
		series.New(jerseyNumbers, series.String, "jersey_number"),
		series.New(positions, series.String, "position"),
		series.New(positionIDs, series.String, "position_id"),
		series.New(homeAways, series.String, "home_away"),
	)

	return df
}
