package releases

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const (
	API_URL = "https://api.github.com/repos/wingnut-tech/LEDController/releases"
)

// Layouts is a map of filename:url ({"radian_v1.x.x": "https://..."})
type Layouts map[string]string

// Versions is a map of vernum:layouts ({"v1.x.x": Layouts(radian, ...)})
type Versions map[string]Layouts

// GithubReleases is a generic container for any github release info
type GithubReleases []struct {
	Name   string `json:"name"`
	Assets []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func ParseReleases(url string) (GithubReleases, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	releases := GithubReleases{}
	err = json.Unmarshal(body, &releases)
	if err != nil {
		return nil, err
	}

	return releases, nil
}

func GetVersions() (Versions, error) {
	versions := make(Versions)
	releases, err := ParseReleases(API_URL)
	if err != nil || len(releases) < 1 {
		return versions, err
	}

	for _, release := range releases {
		versions[release.Name] = Layouts{}
		for _, asset := range release.Assets {
			if strings.HasSuffix(asset.Name, ".hex") {
				versions[release.Name][asset.Name] = asset.URL
			}
		}
	}
	return versions, nil
}
