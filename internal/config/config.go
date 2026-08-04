package config

import (
	"encoding/json"
	"os"
)

// NavLink is a single navbar entry.
type NavLink struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

// Site holds all configurable values for the site.
type Site struct {
	SiteTitle       string    `json:"siteTitle"`
	SiteTitleSuffix string    `json:"siteTitleSuffix"`
	SiteDescription string    `json:"siteDescription"`
	GitHub          string    `json:"github"`
	BaseURL         string    `json:"baseURL"`
	NavLinks        []NavLink `json:"navLinks"`

	// Resolved at build time, not from config.json
	AvatarURL string `json:"-"`
}

// Load reads config.json from the given path and returns a Site.
func Load(path string) (*Site, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var s Site
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}
	return &s, nil
}
