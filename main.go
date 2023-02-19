package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
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
	Playtime2Weeks        int    `json:"playtime_2weeks"`
    PlaytimeForever       int    `json:"playtime_forever"`

}

func main() {

	if os.Getenv("STEAM_TOKEN") == "" {
		panic("Missing STEAM_TOKEN environment variable")
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)

}

func handler(w http.ResponseWriter, r *http.Request) {
    steamid := strings.TrimPrefix(r.URL.Path, "/")

    if steamid == "" || len(steamid) != 17 {
        http.Error(w, "SteamID is incorrect", http.StatusBadRequest)
        return
    }

    // Retrieve Steam profile information
    profileURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s", os.Getenv("STEAM_TOKEN"), steamid)
    profileResp, err := http.Get(profileURL)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer profileResp.Body.Close()

    var steamProfileResponse SteamProfileResponse
    err = json.NewDecoder(profileResp.Body).Decode(&steamProfileResponse)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if len(steamProfileResponse.Response.Players) == 0 {
        http.Error(w, "No player found", http.StatusBadRequest)
        return
    }

    // Retrieve owned games information
    gamesURL := fmt.Sprintf("http://api.steampowered.com/IPlayerService/GetOwnedGames/v1/?key=%s&steamid=%s&format=json&appids_filter[0]=1422130", os.Getenv("STEAM_TOKEN"), steamid)
    gamesResp, err := http.Get(gamesURL)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer gamesResp.Body.Close()

    var gamesResponse struct {
        Response struct {
            GameCount int `json:"game_count"`
            Games     []struct {
                AppID           int `json:"appid"`
                Playtime2Weeks  int `json:"playtime_2weeks"`
                PlaytimeForever int `json:"playtime_forever"`
            } `json:"games"`
        } `json:"response"`
    }
    err = json.NewDecoder(gamesResp.Body).Decode(&gamesResponse)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Add owned games information to player struct
    player := steamProfileResponse.Response.Players[0]
    if len(gamesResponse.Response.Games) > 0 {
        player.Playtime2Weeks = gamesResponse.Response.Games[0].Playtime2Weeks
        player.PlaytimeForever = gamesResponse.Response.Games[0].PlaytimeForever
    }

    w.Header().Set("Cache-Control", "public, maxage=86400")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(player)
}
