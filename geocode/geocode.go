package geocode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

type CoordinatesResults struct {
	Lat, Lng float64
}

type ErrMsg struct {
	Err error
}

type GeoCodingResponse struct {
	Results []struct {
		Geometry struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"geometry"`
		Formatted string `json:"formatted"`
	} `json:"results"`
	Status struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
}

func GetCoordinates(cityName string) (*CoordinatesResults, error) {
	geocodeAPIKey := os.Getenv("OPENCAGE_API_KEY")
	if geocodeAPIKey == "" {
		return nil, fmt.Errorf("OPENCAGE_API_KEY is not set")
	}

	cityName = strings.TrimSpace(cityName)
	q := url.QueryEscape(cityName)
	geocodeAPIURL := fmt.Sprintf("https://api.opencagedata.com/geocode/v1/json?q=%s&key=%s", q, geocodeAPIKey)

	geocodeResp, err := httpClient.Get(geocodeAPIURL)
	if err != nil {
		return nil, fmt.Errorf("unable to reach Geocode API: %w", err)
	}
	defer geocodeResp.Body.Close()

	if geocodeResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocode API returned HTTP: %s", geocodeResp.Status)
	}

	var opencageData GeoCodingResponse
	if err = json.NewDecoder(geocodeResp.Body).Decode(&opencageData); err != nil {
		return nil, fmt.Errorf("failed to decode geocode response body: %w", err)
	}
	if opencageData.Status.Code != 200 {
		return nil, fmt.Errorf("invalid status code: %d - %s", opencageData.Status.Code, opencageData.Status.Message)
	}
	if len(opencageData.Results) == 0 {
		return nil, fmt.Errorf("no coordinates found for %q", cityName)
	}

	return &CoordinatesResults{
		Lat: opencageData.Results[0].Geometry.Lat,
		Lng: opencageData.Results[0].Geometry.Lng,
	}, nil
}

func FetchGeocode(city string) tea.Cmd {

	return func() tea.Msg {
		coords, err := GetCoordinates(city)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return *coords
	}
}
