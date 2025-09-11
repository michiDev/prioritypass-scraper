package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type Airport struct {
	Name      string
	URL       string
	Country   string
	City      string
	Terminals []Terminal
}

type Terminal struct {
	Name        string
	Experiences []Experience
}

type Experience struct {
	Name        string
	URL         string
	Category    string
	Information map[string]string
}

func main() {
	resp, err := http.Get("https://www.prioritypass.com/airport-lounges")
	if err != nil {
		log.Fatalln("http request failed")
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln("failed to read body")
		return
	}
	airports := parseAirports(string(body))
	for i := range airports {
		resp2, err := http.Get("https://www.prioritypass.com" + airports[i].URL)
		if err != nil {
			log.Printf("failed to fetch airport page for %s: %v", airports[i].Name, err)
			continue
		}
		body2, err := io.ReadAll(resp2.Body)
		resp2.Body.Close()
		if err != nil {
			log.Printf("failed to read body for %s: %v", airports[i].Name, err)
			continue
		}
		airports[i].Terminals = parseTerminals(string(body2))
	}
	for i, airport := range airports {
		fmt.Printf("%d. %s\n", i+1, airport.Name)
		fmt.Printf("   Country: %s\n", airport.Country)
		fmt.Printf("   City: %s\n", airport.City)
		fmt.Printf("   URL: %s\n", airport.URL)
		fmt.Printf("   Terminals:\n")
		for j, terminal := range airport.Terminals {
			fmt.Printf("     %d. %s\n", j+1, terminal.Name)
		}
		fmt.Println()
	}
}

func parseAirports(html string) []Airport {
	var airports []Airport
	pattern := `<a class="link-arrow thin[^"]*" href="(/lounges/[^"]+)"[^>]*>\s*([^<]+)<span class="icon-caret-right"></span>\s*</a>`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) >= 3 {
			url := match[1]
			name := strings.TrimSpace(match[2])

			// Skip empty names or URLs that seem to be lounge-specific rather than airport-specific
			if name == "" || strings.Count(url, "/") > 3 {
				continue
			}

			// Extract country and city from URL
			country, city := extractLocationFromURL(url)

			airport := Airport{
				Name:    name,
				URL:     url,
				Country: country,
				City:    city,
			}

			airports = append(airports, airport)
		}
	}
	return airports
}

func parseTerminals(html string) []Terminal {
	var terminals []Terminal
	pattern := `<div class="terminal-name "[^>]*>\s*<div class="terminal-name-text">\s*([^<]+)\s*</div>\s*</div>`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(html, -1)

	for _, match := range matches {
		if len(match) >= 2 {
			name := strings.TrimSpace(match[1])
			terminal := Terminal{
				Name: name,
			}
			terminals = append(terminals, terminal)
		}
	}
	return terminals
}

func extractLocationFromURL(url string) (country, city string) {
	// URL format: /lounges/country/city-airport
	parts := strings.Split(url, "/")
	if len(parts) >= 4 {
		country = strings.ReplaceAll(parts[2], "-", " ")
		country = strings.Title(country)

		city = strings.ReplaceAll(parts[3], "-", " ")
		city = strings.Title(city)
	}
	return country, city
}
