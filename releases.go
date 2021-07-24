package main

import (
	"encoding/json"
	"io"
	"net/http"
	"sort"
	"strings"
)

type Releases []struct {
	Name   string `json:"name"`
	Assets []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func (a *App) UpdateReleases() {
	resp, err := http.Get(API_URL)
	if err != nil {
		a.NewPopup("Error", "Unable to get releases")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		a.NewPopup("Error", "Unable to get releases")
		return
	}

	releases := Releases{}
	err = json.Unmarshal(body, &releases)
	if err != nil {
		a.NewPopup("Error", "Unable to parse releases")
		return
	}

	for _, release := range releases {
		a.releases[release.Name] = map[string]string{}
		for _, asset := range release.Assets {
			if strings.HasSuffix(asset.Name, ".hex") {
				a.releases[release.Name][asset.Name] = asset.BrowserDownloadURL
			}
		}
	}
	a.UpdateVersions()
}

func (a *App) UpdateVersions() {
	o := []string{}
	for k := range a.releases {
		o = append(o, k)
	}

	// sort descending
	sort.Slice(o, func(i, j int) bool {
		return o[i] > o[j]
	})
	a.verSelect.Options = o
	a.verSelect.SetSelectedIndex(0)
}

func (a *App) UpdateLayouts(v string) {
	prevSelection := a.layoutSelect.Selected

	o := []string{}
	for k := range a.releases[v] {
		o = append(o, k)
	}

	// sort descending
	sort.Slice(o, func(i, j int) bool {
		return o[i] > o[j]
	})
	o = append(o, "-Custom-")

	a.layoutSelect.Options = o
	if prevSelection == "-Custom-" {
		a.layoutSelect.SetSelected("-Custom-")
	} else {
		a.layoutSelect.SetSelectedIndex(0)
	}
}
