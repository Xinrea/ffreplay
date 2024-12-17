package markers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const MARKER_ENDPOINT = "https://ffreplay-api-xmwwmrlcfk.cn-hangzhou.fcapp.run"

type WorldMarker struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	Icon  int `json:"icon"`
	MapID int `json:"mapID"`
}

func QueryWorldMarkersFromApi(code string, fight int) []WorldMarker {
	var Response struct {
		Data []WorldMarker
	}
	resp, err := http.Get(fmt.Sprintf("%s/markers/%s/%d", MARKER_ENDPOINT, code, fight))
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&Response)
	if err != nil {
		log.Println(err)
		return nil
	}
	return Response.Data
}

func QueryWorldMarkers(code string, fight int) []WorldMarker {
	var fightResponse struct {
		Fights []struct {
			ID        int `json:"id"`
			StartTime int `json:"start_time"`
			Boss      int `json:"boss"`
		} `json:"fights"`
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.fflogs.com/reports/fights-and-participants/%s/0", code), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("referer", "https://www.fflogs.com/")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&fightResponse)
	if err != nil {
		log.Fatal(err)
	}
	// find boss id in fight
	boss := -1
	startTime := -1
	for _, f := range fightResponse.Fights {
		if f.ID == fight {
			boss = f.Boss
			startTime = f.StartTime
			break
		}
	}
	if boss == -1 {
		log.Fatal("Invalid boss id found")
	}

	req, err = http.NewRequest("GET", fmt.Sprintf("https://www.fflogs.com/reports/replaysegment/%s/%d/%d/%d", code, boss, startTime, startTime), nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("referer", "https://www.fflogs.com/")
	segResp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer segResp.Body.Close()
	var segResponse struct {
		WorldMarkers []WorldMarker `json:"worldMarkers"`
	}
	err = json.NewDecoder(segResp.Body).Decode(&segResponse)
	if err != nil {
		log.Fatal(err)
	}
	return segResponse.WorldMarkers
}
