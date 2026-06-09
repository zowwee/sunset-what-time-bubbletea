package sunset

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type SunsetResponse struct {
	Results struct {
		Sunset string `json:"sunset"`
	} `json:"results"`
	Status string `json:"status"`
	Tzid   string `json:"tzid"`
}

type UTCSunsetTime struct {
	UTCSunsetTime time.Time
}

type ErrMsg struct {
	Err error
}

func GetSunsetTimeInUTC(lat float64, lng float64, cityName string) (*UTCSunsetTime, error) {
	sunsetAPIURL := fmt.Sprintf("https://api.sunrise-sunset.org/json?lat=%v&lng=%v&formatted=0", lat, lng)
	sunsetResp, err := httpClient.Get(sunsetAPIURL)
	if err != nil {
		return nil, fmt.Errorf("unable to reach sunrise-sunset.org API: %w", err)
	}
	defer sunsetResp.Body.Close()

	if sunsetResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sunrise-sunset API returned HTTP %s", sunsetResp.Status)
	}

	var sunsetData SunsetResponse
	if err = json.NewDecoder(sunsetResp.Body).Decode(&sunsetData); err != nil {
		return nil, fmt.Errorf("unable to decode sunrise-sunset.org response: %w", err)
	}

	if sunsetData.Status != "OK" {
		return nil, fmt.Errorf("status: %v", sunsetData.Status)
	}

	if len(sunsetData.Results.Sunset) == 0 {
		return nil, fmt.Errorf("no sunset time found for %q", cityName)
	}

	unformattedSunsetTimeText := sunsetData.Results.Sunset

	unformattedSunsetTime, err := time.Parse(time.RFC3339, unformattedSunsetTimeText)
	if err != nil {
		return nil, fmt.Errorf("unable to format time for %v", cityName)
	}

	return &UTCSunsetTime{
		UTCSunsetTime: unformattedSunsetTime,
	}, nil

}

func FetchSunset(lat float64, lng float64, city string) tea.Cmd {

	return func() tea.Msg {
		result, err := GetSunsetTimeInUTC(lat, lng, city)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return *result
	}

}
