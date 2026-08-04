// Package github fetches public GitHub profile data at build time.
package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var client = &http.Client{Timeout: 10 * time.Second}

type userResponse struct {
	AvatarURL string `json:"avatar_url"`
}

// AvatarURL returns the avatar URL for the given GitHub username.
// Returns an empty string if the username is empty or the request fails.
func AvatarURL(username string) (string, error) {
	if username == "" {
		return "", nil
	}

	resp, err := client.Get(fmt.Sprintf("https://api.github.com/users/%s", username))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api: %s", resp.Status)
	}

	var u userResponse
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return "", err
	}

	return u.AvatarURL, nil
}
