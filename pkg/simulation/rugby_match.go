package simulation

// import (
// 	"fmt"
// 	"math"

// 	"github.com/umbralcalc/stochadex/pkg/simulator"
// 	"golang.org/x/exp/rand"
// 	"gonum.org/v1/gonum/stat/distuv"
// )

// const PitchLon = 100.0
// const PitchLat = 70.0

// var PossessionMap = map[int]string{0: "Home Team", 1: "Away Team"}
// var MatchStateMap = map[int]string{
// 	0:  "Penalty",
// 	1:  "Free Kick",
// 	2:  "Goal",
// 	3:  "Drop Goal",
// 	4:  "Try",
// 	5:  "Kick Phase",
// 	6:  "Run Phase",
// 	7:  "Knock-on",
// 	8:  "Scrum",
// 	9:  "Lineout",
// 	10: "Ruck",
// 	11: "Maul",
// 	12: "Kickoff",
// }
// var MatchStateValueIndices = map[string]int{
// 	"Match State":              0,
// 	"Possession State":         1,
// 	"Longitudinal Pitch State": 2,
// 	"Lateral Pitch State":      3,
// 	"Home Team Score":          4,
// 	"Away Team Score":          5,
// 	"Last Match State":         6,
// 	"Next Match State":         7,
// 	"Current Attacker":         8,
// 	"Current Defender":         9,
// 	"Play Direction":           10,
// }

// // getPlayerFatigue is an internal method to retrieve a player's fatigue factor.
// func getPlayerFatigue(
// 	playerIndex int,
// 	params simulator.Params,
// 	timestepsHistory *simulator.CumulativeTimestepsHistory,
// ) float64 {
// 	return math.Exp(
// 		-params["player_fatigue_rates"][playerIndex] *
// 			(timestepsHistory.Values.AtVec(0) -
// 				params["player_start_times"][playerIndex]),
// 	)
// }

// // getScrumPossessionFactor is an internal method to retrieve the player weightings
// // for the scrum possession transition probability.
// func getScrumPossessionFactor(
// 	state []float64,
// 	params simulator.Params,
// 	timestepsHistory *simulator.CumulativeTimestepsHistory,
// ) float64 {
// 	playersFactor := 0.0
// 	norm := 0.0
// 	for i := 0; i < 3; i++ {
// 		attackingFrontRowPos :=
// 			params["front_row_scrum_possessions"][i+int(3*state[1])] *
// 				getPlayerFatigue(i+int(15*state[1]), params, timestepsHistory)
// 		defendingFrontRowPos :=
// 			params["front_row_scrum_possessions"][i+int(3*(1-state[1]))] *
// 				getPlayerFatigue(i+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingFrontRowPos
// 		norm += attackingFrontRowPos + defendingFrontRowPos
// 	}
// 	for i := 0; i < 2; i++ {
// 		attackingSecondRowPos :=
// 			params["second_row_scrum_possessions"][i+int(2*state[1])] *
// 				getPlayerFatigue(i+3+int(15*state[1]), params, timestepsHistory)
// 		defendingSecondRowPos :=
// 			params["second_row_scrum_possessions"][i+int(2*(1-state[1]))] *
// 				getPlayerFatigue(i+3+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingSecondRowPos
// 		norm += attackingSecondRowPos + defendingSecondRowPos
// 	}
// 	playersFactor /= norm
// 	return playersFactor
// }

// // getLineoutPossessionFactor is an internal method to retrieve the player weightings
// // for the lineout possession transition probability.
// func getLineoutPossessionFactor(
// 	state []float64,
// 	params simulator.Params,
// 	timestepsHistory *simulator.CumulativeTimestepsHistory,
// ) float64 {
// 	playersFactor := 0.0
// 	norm := 0.0
// 	for i := 0; i < 3; i++ {
// 		attackingFrontRowPos :=
// 			params["front_row_lineout_possessions"][i+int(3*state[1])] *
// 				getPlayerFatigue(i+int(15*state[1]), params, timestepsHistory)
// 		defendingFrontRowPos :=
// 			params["front_row_lineout_possessions"][i+int(3*(1-state[1]))] *
// 				getPlayerFatigue(i+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingFrontRowPos
// 		norm += attackingFrontRowPos + defendingFrontRowPos
// 	}
// 	for i := 0; i < 2; i++ {
// 		attackingSecondRowPos :=
// 			params["second_row_lineout_possessions"][i+int(2*state[1])] *
// 				getPlayerFatigue(i+3+int(15*state[1]), params, timestepsHistory)
// 		defendingSecondRowPos :=
// 			params["second_row_lineout_possessions"][i+int(2*(1-state[1]))] *
// 				getPlayerFatigue(i+3+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingSecondRowPos
// 		norm += attackingSecondRowPos + defendingSecondRowPos
// 	}
// 	for i := 0; i < 3; i++ {
// 		attackingBackRowPos :=
// 			params["back_row_lineout_possessions"][i+int(3*state[1])] *
// 				getPlayerFatigue(i+5+int(15*state[1]), params, timestepsHistory)
// 		defendingBackRowPos :=
// 			params["back_row_lineout_possessions"][i+int(3*(1-state[1]))] *
// 				getPlayerFatigue(i+5+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingBackRowPos
// 		norm += attackingBackRowPos + defendingBackRowPos
// 	}
// 	playersFactor /= norm
// 	return playersFactor
// }

// // getMaulPossessionFactor is an internal method to retrieve the player weightings
// // for the maul possession transition probability.
// func getMaulPossessionFactor(
// 	state []float64,
// 	params simulator.Params,
// 	timestepsHistory *simulator.CumulativeTimestepsHistory,
// ) float64 {
// 	playersFactor := 0.0
// 	norm := 0.0
// 	for i := 0; i < 3; i++ {
// 		attackingFrontRowPos :=
// 			params["front_row_maul_possessions"][i+int(3*state[1])] *
// 				getPlayerFatigue(i+int(15*state[1]), params, timestepsHistory)
// 		defendingFrontRowPos :=
// 			params["front_row_maul_possessions"][i+int(3*(1-state[1]))] *
// 				getPlayerFatigue(i+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingFrontRowPos
// 		norm += attackingFrontRowPos + defendingFrontRowPos
// 	}
// 	for i := 0; i < 2; i++ {
// 		attackingSecondRowPos :=
// 			params["second_row_maul_possessions"][i+int(2*state[1])] *
// 				getPlayerFatigue(i+3+int(15*state[1]), params, timestepsHistory)
// 		defendingSecondRowPos :=
// 			params["second_row_maul_possessions"][i+int(2*(1-state[1]))] *
// 				getPlayerFatigue(i+3+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingSecondRowPos
// 		norm += attackingSecondRowPos + defendingSecondRowPos
// 	}
// 	for i := 0; i < 3; i++ {
// 		attackingBackRowPos :=
// 			params["back_row_maul_possessions"][i+int(3*state[1])] *
// 				getPlayerFatigue(i+5+int(15*state[1]), params, timestepsHistory)
// 		defendingBackRowPos :=
// 			params["back_row_maul_possessions"][i+int(3*(1-state[1]))] *
// 				getPlayerFatigue(i+5+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingBackRowPos
// 		norm += attackingBackRowPos + defendingBackRowPos
// 	}
// 	playersFactor /= norm
// 	return playersFactor
// }

// // getRuckPossessionFactor is an internal method to retrieve the player weightings
// // for the ruck possession transition probability.
// func getRuckPossessionFactor(
// 	state []float64,
// 	params simulator.Params,
// 	timestepsHistory *simulator.CumulativeTimestepsHistory,
// ) float64 {
// 	playersFactor := 0.0
// 	norm := 0.0
// 	for i := 0; i < 3; i++ {
// 		attackingFrontRowPos :=
// 			params["front_row_ruck_possessions"][i+int(3*state[1])] *
// 				getPlayerFatigue(i+int(15*state[1]), params, timestepsHistory)
// 		defendingFrontRowPos :=
// 			params["front_row_ruck_possessions"][i+int(3*(1-state[1]))] *
// 				getPlayerFatigue(i+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingFrontRowPos
// 		norm += attackingFrontRowPos + defendingFrontRowPos
// 	}
// 	for i := 0; i < 2; i++ {
// 		attackingSecondRowPos :=
// 			params["second_row_ruck_possessions"][i+int(2*state[1])] *
// 				getPlayerFatigue(i+3+int(15*state[1]), params, timestepsHistory)
// 		defendingSecondRowPos :=
// 			params["second_row_ruck_possessions"][i+int(2*(1-state[1]))] *
// 				getPlayerFatigue(i+3+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingSecondRowPos
// 		norm += attackingSecondRowPos + defendingSecondRowPos
// 	}
// 	for i := 0; i < 3; i++ {
// 		attackingBackRowPos :=
// 			params["back_row_ruck_possessions"][i+int(3*state[1])] *
// 				getPlayerFatigue(i+5+int(15*state[1]), params, timestepsHistory)
// 		defendingBackRowPos :=
// 			params["back_row_ruck_possessions"][i+int(3*(1-state[1]))] *
// 				getPlayerFatigue(i+5+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingBackRowPos
// 		norm += attackingBackRowPos + defendingBackRowPos
// 	}
// 	for i := 0; i < 2; i++ {
// 		attackingCentresPos :=
// 			params["centres_ruck_possessions"][i+int(2*state[1])] *
// 				getPlayerFatigue(i+11+int(15*state[1]), params, timestepsHistory)
// 		defendingCentresPos :=
// 			params["centres_ruck_possessions"][i+int(2*(1-state[1]))] *
// 				getPlayerFatigue(i+11+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingCentresPos
// 		norm += attackingCentresPos + defendingCentresPos
// 	}
// 	playersFactor /= norm
// 	return playersFactor
// }

// // getRunPossessionFactor is an internal method to retrieve the player weightings
// // for the run possession transition probability.
// func getRunPossessionFactor(
// 	state []float64,
// 	params simulator.Params,
// 	timestepsHistory *simulator.CumulativeTimestepsHistory,
// ) float64 {
// 	playersFactor := 0.0
// 	norm := 0.0
// 	for i := 0; i < 15; i++ {
// 		attackingPos :=
// 			params["player_run_possessions"][i+int(3*state[1])] *
// 				getPlayerFatigue(i+int(15*state[1]), params, timestepsHistory)
// 		defendingPos :=
// 			params["player_run_possessions"][i+int(3*(1-state[1]))] *
// 				getPlayerFatigue(i+int(15*(1-state[1])), params, timestepsHistory)
// 		playersFactor += defendingPos
// 		norm += attackingPos + defendingPos
// 	}
// 	playersFactor /= norm
// 	return playersFactor
// }

// // RugbyMatchIteration defines an iteration for a model of a rugby match
// // defined in this article: https://umbralcalc.github.io/posts/trywizard.html
// type RugbyMatchIteration struct {
// 	maxLat          float64
// 	maxLon          float64
// 	indices         map[string]int
// 	normalDist      *distuv.Normal
// 	unitUniformDist *distuv.Uniform
// 	exponentialDist *distuv.Exponential
// 	categoricalDist *distuv.Categorical
// }

// func (r *RugbyMatchIteration) getPossessionChange(
// 	state []float64,
// 	params simulator.Params,
// 	timestepsHistory *simulator.CumulativeTimestepsHistory,
// ) []float64 {
// 	rate := params["max_possession_change_rates"][int(state[0])]
// 	playersFactor := 1.0
// 	switch state[0] {
// 	case 6:
// 		playersFactor = getRunPossessionFactor(state, params, timestepsHistory)
// 	case 8:
// 		playersFactor = getScrumPossessionFactor(state, params, timestepsHistory)
// 	case 9:
// 		playersFactor = getLineoutPossessionFactor(state, params, timestepsHistory)
// 	case 10:
// 		playersFactor = getRuckPossessionFactor(state, params, timestepsHistory)
// 	case 11:
// 		playersFactor = getMaulPossessionFactor(state, params, timestepsHistory)
// 	}
// 	rate *= playersFactor
// 	if rate > (rate+(1.0/timestepsHistory.NextIncrement))*r.unitUniformDist.Rand() {
// 		state[1] = (1.0 - state[1])
// 	}
// 	return state
// }

// func (r *RugbyMatchIteration) getLongitudinalRunChange(
// 	state []float64,
// 	params simulator.Params,
// ) []float64 {
// 	newLonState := state[2]
// 	attackerIndex := int(state[r.indices["Current Attacker"]]) + int(15*state[1])
// 	defenderIndex := int(state[r.indices["Current Defender"]]) + int(15*(1-state[1]))
// 	r.exponentialDist.Rate =
// 		params["player_defensive_run_scales"][defenderIndex]
// 	newLonState -= state[r.indices["Play Direction"]] * r.exponentialDist.Rand()
// 	r.exponentialDist.Rate =
// 		params["player_attacking_run_scales"][attackerIndex]
// 	newLonState += state[r.indices["Play Direction"]] * r.exponentialDist.Rand()
// 	// if the newLonState would end up moving over a tryline, just restrict
// 	// this movement so that it remains just within the field of play
// 	if newLonState > r.maxLon {
// 		newLonState = r.maxLon - 0.5
// 	}
// 	if newLonState < 0.0 {
// 		newLonState = 0.5
// 	}
// 	state[2] = newLonState
// 	return state
// }

// func (r *RugbyMatchIteration) getLateralRunChange(
// 	state []float64,
// 	params simulator.Params,
// ) []float64 {
// 	r.normalDist.Mu = 0.0
// 	r.normalDist.Sigma = params["lateral_run_scale"][0]
// 	newLatState := state[3] + r.normalDist.Rand()
// 	// if the newLatState would end up moving out of bounds, just restrict
// 	// this movement so that it remains just within the field of play
// 	if newLatState > r.maxLat {
// 		newLatState = r.maxLat - 0.5
// 	}
// 	if newLatState < 0.0 {
// 		newLatState = 0.5
// 	}
// 	state[3] = newLatState
// 	return state
// }

// func (r *RugbyMatchIteration) getLongitudinalKickChange(
// 	state []float64,
// 	params simulator.Params,
// ) []float64 {
// 	lastState := int(state[r.indices["Last Match State"]])
// 	// if this is a kick at goal or a drop goal don't move
// 	if ((lastState == 0) && (state[0] == 2)) ||
// 		((lastState == 5) && (state[0] == 3)) {
// 		return state
// 	}
// 	// if this is a kick in the field of play
// 	if (lastState == 5) && (state[0] != 3) && (state[0] != 9) {
// 		newLonState := state[2]
// 		// choose a kicker at random
// 		possibleKickers := []float64{9, 10, 11, 14, 15}
// 		currentAttacker := int(possibleKickers[int(rand.Intn(5))])
// 		var kickerIndex int
// 		if (currentAttacker == 9) || (currentAttacker == 10) {
// 			kickerIndex = (currentAttacker - 9) + 2*int(state[1])
// 			r.exponentialDist.Rate =
// 				params["halves_kick_scales"][kickerIndex]
// 		} else {
// 			if currentAttacker == 11 {
// 				kickerIndex = 3 * int(state[1])
// 			} else {
// 				kickerIndex = (currentAttacker - 13) + 3*int(state[1])
// 			}
// 			r.exponentialDist.Rate =
// 				params["back_three_kick_scales"][kickerIndex]
// 		}
// 		newLonState += state[r.indices["Play Direction"]] * r.exponentialDist.Rand()
// 		if newLonState >= r.maxLon {
// 			newLonState = r.maxLon - 0.5
// 		}
// 		state[2] = newLonState
// 		return state
// 	}
// 	// if this is a kick to touch
// 	if (lastState == 5) && (state[0] == 9) {
// 		return state
// 	}
// 	return state
// }

// func (r *RugbyMatchIteration) getLateralKickChange(
// 	state []float64,
// 	_ simulator.Params,
// ) []float64 {
// 	lastState := int(state[r.indices["Last Match State"]])
// 	// if this is a kick at goal or a drop goal don't move
// 	if ((lastState == 0) && (state[0] == 2)) ||
// 		((lastState == 5) && (state[0] == 3)) {
// 		return state
// 	}
// 	// if this is a kick in the field of play
// 	if (lastState == 5) && (state[0] != 3) && (state[0] != 9) {
// 		state[3] = r.maxLat * r.unitUniformDist.Rand()
// 		return state
// 	}
// 	// if this is a kick to touch
// 	if (lastState == 5) && (state[0] == 9) {
// 		if r.unitUniformDist.Rand() > 0.5 {
// 			state[3] = r.maxLat
// 		} else {
// 			state[3] = 0.0
// 		}
// 	}
// 	return state
// }

// func (r *RugbyMatchIteration) getKickAtGoalSuccess(
// 	state []float64,
// 	params simulator.Params,
// ) bool {
// 	success := r.unitUniformDist.Rand() <
// 		params["goal_probabilities"][int(state[1])]
// 	midPitch := 0.5 * r.maxLat
// 	if success {
// 		// move ball back to halfway line for kickoff
// 		line50m := 0.5 * r.maxLon
// 		state[2] = line50m
// 		state[3] = midPitch
// 	} else {
// 		// move ball to 22 for a dropout (another kind of kickoff event)
// 		line22m := r.maxLon * (0.78*state[1] + 0.22*(1-state[1]))
// 		state[2] = line22m
// 		state[3] = midPitch
// 	}
// 	return success
// }

// func (r *RugbyMatchIteration) updateScoreAndBallLocation(
// 	state []float64,
// 	params simulator.Params,
// ) []float64 {
// 	// update either home team or away team scores with this index
// 	scorerIndex := int(5*state[1] + 4*(1-state[1]))
// 	line22m := r.maxLon * (0.78*state[1] + 0.22*(1-state[1]))
// 	midPitch := 0.5 * r.maxLat
// 	// update home team score
// 	if (state[0] == 2) || (state[0] == 3) {
// 		if r.getKickAtGoalSuccess(state, params) {
// 			state[scorerIndex] += 3.0
// 		} else {
// 			// if unsuccessful with a penalty or drop goal, restart with dropout
// 			state[2] = line22m
// 			state[3] = midPitch
// 		}
// 	} else if state[0] == 4 {
// 		state[scorerIndex] += 5.0
// 		// always by default move the ball back to 22m line after a try is
// 		// scored, ready to kick at goal
// 		state[2] = line22m
// 		if r.getKickAtGoalSuccess(state, params) {
// 			state[scorerIndex] += 2.0
// 		}
// 	}
// 	return state
// }

// func (r *RugbyMatchIteration) possessionChangeCanOccur(state []float64) bool {
// 	cantOccur := []float64{0, 1, 2, 3, 4, 7, 12}
// 	for _, value := range cantOccur {
// 		if value == state[0] {
// 			return false
// 		}
// 	}
// 	return true
// }

// func (r *RugbyMatchIteration) Configure(
// 	partitionIndex int,
// 	settings *simulator.Settings,
// ) {
// 	seed := settings.Seeds[partitionIndex]
// 	weights := make([]float64, 0)
// 	for i := 0; i < 15; i++ {
// 		weights = append(weights, 1.0)
// 	}
// 	catDist := distuv.NewCategorical(weights, rand.NewSource(seed))
// 	rand.Seed(seed)

// 	r.indices = MatchStateValueIndices
// 	r.maxLon = PitchLon
// 	r.maxLat = PitchLat
// 	r.normalDist = &distuv.Normal{
// 		Mu:    0.0,
// 		Sigma: 1.0,
// 		Src:   rand.NewSource(seed),
// 	}
// 	r.unitUniformDist = &distuv.Uniform{
// 		Min: 0.0,
// 		Max: 1.0,
// 		Src: rand.NewSource(seed),
// 	}
// 	r.exponentialDist = &distuv.Exponential{
// 		Rate: 1.0,
// 		Src:  rand.NewSource(seed),
// 	}
// 	r.categoricalDist = &catDist
// }

// func (r *RugbyMatchIteration) Iterate(
// 	params simulator.Params,
// 	partitionIndex int,
// 	stateHistories []*simulator.StateHistory,
// 	timestepsHistory *simulator.CumulativeTimestepsHistory,
// ) []float64 {
// 	state := make([]float64, 0)
// 	state = append(state, stateHistories[partitionIndex].Values.RawRowView(0)...)
// 	state[r.indices["Play Direction"]] = 1.0*state[1] - 1.0*(1-state[1])
// 	matchState := fmt.Sprintf("%d", int(state[0]))
// 	transitions := params["transitions_from_"+matchState]
// 	// if we are currently not planned to do anything, find the next transition
// 	if state[0] == float64(state[r.indices["Next Match State"]]) {
// 		// compute the cumulative rates and overall normalisation for transitions
// 		cumulative := 0.0
// 		cumulativeProbs := make([]float64, 0)
// 		transitionProbs := params["transition_probs_from_"+matchState]
// 		for _, prob := range transitionProbs {
// 			cumulative += prob
// 			cumulativeProbs = append(cumulativeProbs, cumulative)
// 		}
// 		normalisation := cumulativeProbs[len(cumulativeProbs)-1]
// 		transitionEvent := r.unitUniformDist.Rand()
// 		for i, prob := range cumulativeProbs {
// 			if transitionEvent*normalisation < prob {
// 				if (i == 0) || (transitionEvent*normalisation >= cumulativeProbs[i-1]) {
// 					state[r.indices["Next Match State"]] = transitions[i]
// 					break
// 				}
// 			}
// 		}
// 	}
// 	// figure out if the next event should happen yet
// 	probDoNothing := 1.0 / (1.0 + timestepsHistory.NextIncrement*
// 		params["background_event_rates"][int64(state[r.indices["Next Match State"]])])
// 	event := r.unitUniformDist.Rand()
// 	if event < probDoNothing {
// 		// if the state hasn't changed then continue without doing anything else
// 		return state
// 	} else {
// 		// else change the state
// 		state[r.indices["Last Match State"]] = state[0]
// 		state[0] = state[r.indices["Next Match State"]]
// 	}
// 	// if at kickoff, reset the ball location to be central and continue
// 	if state[0] == 12 {
// 		state[2] = 0.5 * r.maxLon
// 		state[3] = 0.5 * r.maxLat
// 		return state
// 	}
// 	// if a knock-on has led to a scrum, change possession and continue
// 	if (int(state[r.indices["Last Match State"]]) == 7) && (state[0] == 8) {
// 		state[1] = (1 - state[1])
// 		return state
// 	}
// 	// randomly select new attacking and defending player indices
// 	state[r.indices["Current Attacker"]] = r.categoricalDist.Rand()
// 	state[r.indices["Current Defender"]] = r.categoricalDist.Rand()
// 	// handle scoring if there was a score event and then continue
// 	if (state[0] == 2) || (state[0] == 3) || (state[0] == 4) {
// 		state = r.updateScoreAndBallLocation(state, params)
// 		return state
// 	}
// 	// find out if there is a change in possession if possible
// 	if r.possessionChangeCanOccur(state) {
// 		state = r.getPossessionChange(state, params, timestepsHistory)
// 	}
// 	// if the next phase is a run phase and we are entering this for the first time
// 	// then decide on what spatial movements the ball location makes as a result
// 	if state[0] == 6 {
// 		state = r.getLongitudinalRunChange(state, params)
// 		state = r.getLateralRunChange(state, params)
// 	}
// 	// similarly, if the next phase is a kick phase and we are entering this for
// 	// the first time then decide on what spatial movements the ball location makes
// 	if state[0] == 5 {
// 		state = r.getLongitudinalKickChange(state, params)
// 		state = r.getLateralKickChange(state, params)
// 	}
// 	return state
// }
