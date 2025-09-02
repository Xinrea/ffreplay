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

func queryBossIDAndStart(code string, fight int) (int, int) {
	var fightResponse struct {
		Fights []struct {
			ID        int `json:"id"`
			StartTime int `json:"start_time"`
			Boss      int `json:"boss"`
		} `json:"fights"`
	}

	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("https://www.fflogs.com/reports/fights-and-participants/%s/0", code),
		nil,
	)
	if err != nil {
		log.Panic(err)
	}

	req.Header.Set("referer", fmt.Sprintf("https://www.fflogs.com/reports/%s?fight=%d", code, fight))
	req.Header.Set("user-agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)

		return -1, -1
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&fightResponse)
	if err != nil {
		log.Panic(err)
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
		log.Panic("Invalid boss id found")
	}

	return boss, startTime
}

func QueryWorldMarkers(code string, fight int) []WorldMarker {
	boss, startTime := queryBossIDAndStart(code, fight)
	if boss == -1 {
		return nil
	}

	req, err := http.NewRequest(http.MethodGet,
		fmt.Sprintf("https://www.fflogs.com/reports/replaysegment/%s/%d/%d/%d", code, boss, startTime, startTime),
		nil,
	)
	if err != nil {
		log.Panic(err)
	}

	req.Header.Set("referer", "https://www.fflogs.com/")
	req.Header.Set("user-agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/139.0.0.0 Safari/537.36")

	segResp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)

		return nil
	}
	defer segResp.Body.Close()

	var segResponse struct {
		WorldMarkers []WorldMarker `json:"worldMarkers"`
	}

	log.Println(segResp.Body)

	err = json.NewDecoder(segResp.Body).Decode(&segResponse)
	if err != nil {
		log.Panic(err)
	}

	return segResponse.WorldMarkers
}
