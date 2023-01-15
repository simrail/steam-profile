package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type SteamProfileResponse struct {
	Response SteamProfilePlayerList `json:"response"`
}

type SteamProfilePlayerList struct {
	Players []SteamProfilePlayer `json:"players"`
}

type SteamProfilePlayer struct {
	SteamID             string `json:"steamid"`
	VisibilityState     int    `json:"communityvisibilitystate"`
	ProfileState        int    `json:"profilestate"`
	PersonaName         string `json:"personaname"`
	ProfileURL          string `json:"profileurl"`
	AvatarURL           string `json:"avatar"`
	AvatarMediumURL     string `json:"avatarmedium"`
	AvatarFullURL       string `json:"avatarfull"`
	AvatarHash          string `json:"avatarhash"`
	PersonaState        int    `json:"personastate"`
	RealName            string `json:"realname"`
	PrimaryClanID       string `json:"primaryclanid"`
	TimeCreated         int    `json:"timecreated"`
	PersonaStateFlags   int    `json:"personastateflags"`
	LocationCountryCode string `json:"loccountrycode"`
}

func main() {

	if os.Getenv("STEAM_TOKEN") == "" {
		panic("Missing STEAM_TOKEN environment variable")
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)

}

func handler(w http.ResponseWriter, r *http.Request) {
	steamid := r.URL.Query().Get("steamid")

	if steamid == "" || len(steamid) != 17 {
		http.Error(w, "SteamID is incorrect", http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s", os.Getenv("STEAM_TOKEN"), steamid)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var steamProfileResponse SteamProfileResponse
	err = json.NewDecoder(resp.Body).Decode(&steamProfileResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(steamProfileResponse.Response.Players) == 0 {
		http.Error(w, "No player found", http.StatusBadRequest)
		return
	}

	player := steamProfileResponse.Response.Players[0]

	w.Header().Set("Cache-Control", "public, maxage=86400")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(player)
}
