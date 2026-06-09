package main

import (
	"fmt"
	"log"
	"sunset-what-time-bubbletea/geocode"
	"sunset-what-time-bubbletea/sunset"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/sahilm/fuzzy"
	tzm "github.com/zsefvlol/timezonemapper"
)

type model struct {
	cities []string

	enteringCustomCity bool
	input              string

	submitted bool

	city        string
	cursor      int
	fuzzyCursor int

	coordinates geocode.CoordinatesResults
	sunsetTime  struct {
		utcSunsetTime   time.Time
		localSunsetTime time.Time
	}
	timezone string
	err      error
}

func initialModel() model {
	return model{
		cities: []string{
			"Singapore", "Stockholm", "Newcastle Upon Tyne", "Others",
		},
		fuzzyCursor: -1,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:

		if !m.submitted {
			if m.enteringCustomCity {
				switch msg.String() {

				case "ctrl+c":
					return m, tea.Quit

				case "up":
					if m.fuzzyCursor >= 0 {
						m.fuzzyCursor--
					}

				case "down":
					results := fuzzy.Find(m.input, WorldCities)
					max := len(results)
					if max > 5 {
						max = 5
					}
					if m.fuzzyCursor < max-1 {
						m.fuzzyCursor++
					}

				case "enter":
					if m.input == "" {
						return m, nil
					}

					if m.fuzzyCursor == -1 {
						m.city = m.input
					} else {
						results := fuzzy.Find(m.input, WorldCities)
						m.city = results[m.fuzzyCursor].Str
					}

					m.input = ""
					m.enteringCustomCity = false
					m.submitted = true

					return m, geocode.FetchGeocode(m.city)

				case "backspace":
					if m.input != "" {
						m.input = m.input[:len(m.input)-1]
						m.fuzzyCursor = 0
					}

				default:
					if len(msg.String()) == 1 {
						m.input += msg.String()
						m.fuzzyCursor = -1
					}
				}
				return m, nil
			}

			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "up":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down":
				if m.cursor < len(m.cities)-1 {
					m.cursor++
				}
			case "enter":
				selected := m.cities[m.cursor]
				if selected == "Others" {
					m.enteringCustomCity = true
					return m, nil
				}
				m.city = selected
				m.submitted = true
				return m, geocode.FetchGeocode(m.city)
			}
			return m, nil
		}

		if m.submitted {
			switch msg.String() {
			case "enter":
				m.submitted = !m.submitted
				m = initialModel()
			case "ctrl+c", "q":
				return m, tea.Quit
			}
		}
	case geocode.CoordinatesResults:
		m.coordinates = msg
		m.timezone = getTimezone(m.coordinates.Lat, m.coordinates.Lng)
		return m, sunset.FetchSunset(m.coordinates.Lat, m.coordinates.Lng, m.city)
	case geocode.ErrMsg:
		m.err = msg.Err
		return m, nil
	case sunset.UTCSunsetTime:
		m.sunsetTime.utcSunsetTime = msg.UTCSunsetTime
		if localSunsetTime, err := getLocalSunsetTime(m.sunsetTime.utcSunsetTime, m.timezone); err == nil {
			m.sunsetTime.localSunsetTime = localSunsetTime
		}
		return m, nil
	case sunset.ErrMsg:
		m.err = msg.Err
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	help := HelpStyle.Render("ctrl+c to quit")

	// --- Custom city input ---
	if m.enteringCustomCity {
		s := TitleStyle.Render("Enter a city:") + " " + m.input + "\n\n"

		if m.input != "" {
			results := fuzzy.Find(m.input, WorldCities)
			for i, r := range results {
				if i >= 5 {
					break
				}
				line := "   " + r.Str
				if m.fuzzyCursor == i {
					line = SelectedStyle.Render(" > " + r.Str)
				}
				s += line + "\n"
			}
			if len(results) > 0 {
				s += "\n" + HelpStyle.Render("up/down to navigate, enter to select") + "\n"
			}
		}

		s += "\n" + help
		return s
	}

	// --- City menu ---
	if m.city == "" {
		s := TitleStyle.Render("Select a city:") + "\n\n"
		for i, city := range m.cities {
			line := "   " + city
			if m.cursor == i {
				line = SelectedStyle.Render(" > " + city)
			}
			s += line + "\n"
		}
		s += "\n" + help
		return s
	}

	// --- Error ---
	if m.err != nil {
		return BoxStyle.Render(
			ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n" +
				HelpStyle.Render("enter to try another city · q to quit"),
		)
	}

	// --- Loading states ---
	if m.coordinates.Lat == 0 && m.coordinates.Lng == 0 {
		return BoxStyle.Render(
			fmt.Sprintf("Loading %s...\n\n", TitleStyle.Render(m.city)) +
				HelpStyle.Render("ctrl+c to quit"),
		)
	}

	if m.timezone == "" {
		return BoxStyle.Render(
			fmt.Sprintf("No timezone found for %s.\n\n", TitleStyle.Render(m.city)) +
				HelpStyle.Render("enter to try another city · ctrl+c to quit"),
		)
	}

	if m.sunsetTime.utcSunsetTime.IsZero() {
		return BoxStyle.Render(
			fmt.Sprintf("Fetching sunset time for %s...\n\n", TitleStyle.Render(m.city)) +
				HelpStyle.Render("ctrl+c to quit"),
		)
	}

	// --- Final result ---
	if m.sunsetTime.localSunsetTime.IsZero() {
		formattedUTC := m.sunsetTime.utcSunsetTime.Format("15:04")
		return BoxStyle.Render(
			fmt.Sprintf("The sun sets at %s in %s (UTC)\n\n",
				SelectedStyle.Render(formattedUTC),
				TitleStyle.Render(m.city)) +
				HelpStyle.Render("enter for another city · ctrl+c to quit"),
		)
	}

	formattedLocal := m.sunsetTime.localSunsetTime.Format("15:04")
	return BoxStyle.Render(
		TitleStyle.Render("🌅 "+m.city) + "\n\n" +
			"Sunset: " + SelectedStyle.Render(formattedLocal) + "\n" +
			HelpStyle.Render(m.timezone) + "\n\n" +
			HelpStyle.Render("enter for another city · ctrl+c to quit"),
	)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	if err != nil {
		fmt.Println("Error:", err)
	}
}

func getTimezone(lat, lng float64) string {
	tz := tzm.LatLngToTimezoneString(lat, lng)
	return tz
}

func getLocalSunsetTime(utcSunsetTime time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	localSunsetTime := utcSunsetTime.In(loc)
	return localSunsetTime, nil
}
