package espn

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Summary struct {
	Header Header `json:"header"`
}

type Header struct {
	ID           string        `json:"id"`
	UID          string        `json:"uid"`
	Season       Season        `json:"season"`
	TimeValid    bool          `json:"timeValid"`
	Competitions []Competition `json:"competitions"`
}

type Season struct {
	Year int `json:"year"`
	Type int `json:"type"`
}

type Competition struct {
	ID                    string       `json:"id"`
	UID                   string       `json:"uid"`
	Date                  string       `json:"date"`
	NeutralSite           bool         `json:"neutralSite"`
	ConferenceCompetition bool         `json:"conferenceCompetition"`
	BoxscoreAvailable     bool         `json:"boxscoreAvailable"`
	CommentaryAvailable   bool         `json:"commentaryAvailable"`
	LiveAvailable         bool         `json:"liveAvailable"`
	OnWatchESPN           bool         `json:"onWatchESPN"`
	Recent                bool         `json:"recent"`
	WallclockAvailable    bool         `json:"wallclockAvailable"`
	BoxscoreSource        string       `json:"boxscoreSource"`
	PlayByPlaySource      string       `json:"playByPlaySource"`
	Competitors           []Competitor `json:"competitors"`
	Status                Status       `json:"status"`
	Details               []Event      `json:"details"`
}

type Competitor struct {
	ID         string      `json:"id"`
	UID        string      `json:"uid"`
	Order      int         `json:"order"`
	HomeAway   string      `json:"homeAway"`
	Winner     bool        `json:"winner"`
	Team       Team        `json:"team"`
	Score      string      `json:"score"`
	Linescores []Linescore `json:"linescores"`
	Record     []Record    `json:"record"`
	Possession bool        `json:"possession"`
	Form       string      `json:"form"`
}

type Team struct {
	ID               string `json:"id"`
	GUID             string `json:"guid"`
	UID              string `json:"uid"`
	Name             string `json:"name"`
	Abbreviation     string `json:"abbreviation"`
	DisplayName      string `json:"displayName"`
	ShortDisplayName string `json:"shortDisplayName"`
	Color            string `json:"color"`
	Logos            []Logo `json:"logos"`
	Links            []Link `json:"links"`
}

type Logo struct {
	Href        string   `json:"href"`
	Width       int      `json:"width"`
	Height      int      `json:"height"`
	Alt         string   `json:"alt"`
	Rel         []string `json:"rel"`
	LastUpdated string   `json:"lastUpdated"`
}

type Link struct {
	Rel  []string `json:"rel"`
	Href string   `json:"href"`
	Text string   `json:"text"`
}

type Linescore struct {
	DisplayValue string `json:"displayValue"`
}

type Record struct {
	Type         string `json:"type"`
	Summary      string `json:"summary"`
	DisplayValue string `json:"displayValue"`
}

type Status struct {
	Type StatusType `json:"type"`
}

type StatusType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	State       string `json:"state"`
	Completed   bool   `json:"completed"`
	Description string `json:"description"`
	Detail      string `json:"detail"`
	ShortDetail string `json:"shortDetail"`
}

type Event struct {
	SequenceNumber string        `json:"sequenceNumber"`
	Type           EventType     `json:"type"`
	AwayScore      int           `json:"awayScore"`
	HomeScore      int           `json:"homeScore"`
	Period         Period        `json:"period"`
	Clock          Clock         `json:"clock"`
	AddedClock     Clock         `json:"addedClock"`
	Team           Team          `json:"team"`
	Participants   []Participant `json:"participants"`
}

type EventType struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type Period struct {
	Number int `json:"number"`
}

type Clock struct {
	Value        float64 `json:"value"`
	DisplayValue string  `json:"displayValue"`
}

type Participant struct {
	Athlete Athlete `json:"athlete"`
}

type Athlete struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

func FetchEvents(gameID string) ([]Event, error) {
	url := fmt.Sprintf("https://site.api.espn.com/apis/site/v2/sports/rugby/267979/summary?event=%s", gameID)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var sum Summary
	if err := json.Unmarshal(body, &sum); err != nil {
		return nil, err
	}
	return sum.Header.Competitions[0].Details, nil
}
