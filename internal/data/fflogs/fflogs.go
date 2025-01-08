package fflogs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Xinrea/ffreplay/internal/data/markers"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/util"
)

const ENDPOINT = "https://cn.fflogs.com/api/v2/client"

type FFLogsClient struct {
	client *http.Client
}

func NewFFLogsClient(clientID, clientSecret string) *FFLogsClient {
	creds, err := getFFLogsToken(clientID, clientSecret)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	// create http client with bearer token
	httpClient := &http.Client{
		Transport: &BearerAuthTransport{
			Token: creds.AccessToken,
		},
	}

	return &FFLogsClient{
		client: httpClient,
	}
}

// Have to do this, as fflogs graphql not providing ways to query worldmarkers.
func (c *FFLogsClient) QueryWorldMarkers(code string, fight int) []markers.WorldMarker {
	if util.IsWasm() {
		return markers.QueryWorldMarkersFromApi(code, fight)
	}

	return markers.QueryWorldMarkers(code, fight)
}

func (c *FFLogsClient) RawQuery(query string, variables map[string]any, result any) error {
	for k, v := range variables {
		query = strings.ReplaceAll(query, "$"+k, fmt.Sprintf("%v", v))
	}

	requestBody, err := json.Marshal(map[string]string{
		"query": query,
	})
	if err != nil {
		return err
	}

	resp, err := c.client.Post(ENDPOINT, "application/json", bytes.NewReader(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &result)

	return err
}

func (c *FFLogsClient) QueryMapInfo(mapCode int) GameMap {
	var Query struct {
		Data struct {
			GameData struct {
				Map GameMap
			}
		}
	}

	variables := map[string]interface{}{
		"id": mapCode,
	}

	err := c.RawQuery(`
			query {
				gameData {
					map(id: $id) {
						id
						filename
						sizeFactor
						offsetX
						offsetY
					}
				}
			}
		`, variables, &Query)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(mapCode, Query)

	return Query.Data.GameData.Map
}

func (c *FFLogsClient) QueryActors(reportCode string) []Actor {
	var Query struct {
		Data struct {
			ReportData struct {
				Report struct {
					MasterData struct {
						Actors []Actor
					}
				}
			}
		}
	}

	variables := map[string]interface{}{
		"code": reportCode,
	}

	err := c.RawQuery(`
			query {
				reportData {
					report(code: "$code") {
						masterData {
							actors {
								gameID
								id
								name
								subType
							}
						}
					}
				}
			}
		`, variables, &Query)
	if err != nil {
		log.Fatal(err)
	}

	return Query.Data.ReportData.Report.MasterData.Actors
}

func (c *FFLogsClient) QueryReportFights(reportCode string) []ReportFight {
	var Query struct {
		Data struct {
			ReportData struct {
				Report struct {
					Fights []ReportFight
				}
			}
		}
	}

	variables := map[string]interface{}{
		"code": reportCode,
	}

	err := c.RawQuery(`
			query {
				reportData {
					report(code: "$code") {
						fights {
							id
							name
							startTime
							endTime
							enemyNPCs {
								gameID
								id
								instanceCount
							}
							friendlyPets {
								gameID
								id
								instanceCount
							}
							maps {
								id
							}
							phaseTransitions {
								id
								startTime
							}
						}
					}
				}
			}
		`, variables, &Query)
	if err != nil {
		log.Println(err)

		return nil
	}

	return Query.Data.ReportData.Report.Fights
}

func (c *FFLogsClient) QueryFightPlayers(reportCode string, fightID int) *PlayerDetails {
	var Query struct {
		Data struct {
			ReportData struct {
				Report struct {
					PlayerDetails json.RawMessage
				}
			}
		}
	}

	variables := map[string]interface{}{
		"code":     reportCode,
		"fightIDs": []int{fightID},
	}

	err := c.RawQuery(`
			query {
				reportData {
					report(code: "$code") {
						playerDetails(fightIDs: $fightIDs)
					}
				}
			}
		`, variables, &Query)
	if err != nil {
		log.Println(err)

		return nil
	}

	var players struct {
		Data struct {
			PlayerDetails PlayerDetails `json:"playerDetails"`
		}
	}

	err = json.Unmarshal(Query.Data.ReportData.Report.PlayerDetails, &players)
	if err != nil {
		log.Println(err)

		return nil
	}

	return &players.Data.PlayerDetails
}

type ReportEventPaginator struct {
	Data              json.RawMessage
	NextPageTimestamp float64
}

const RawQueryFightEvents = `
query {
	reportData {
		report(code: "$code") {
			events(
				fightIDs: $fightIDs,
				startTime: $startTime,
				endTime: $endTime,
				limit: 10000,
				includeResources: true,
				useAbilityIDs: false
			) {
				data
				nextPageTimestamp
			}
		}
	}
}
`

func (c *FFLogsClient) QueryFightEvents(reportCode string, fight ReportFight) (ret []FFLogsEvent) {
	var Query struct {
		Data struct {
			ReportData struct {
				Report struct {
					Events ReportEventPaginator
				}
			}
		}
	}

	variables := map[string]interface{}{
		"code":      reportCode,
		"fightIDs":  []int{fight.ID},
		"startTime": fight.StartTime,
		"endTime":   fight.EndTime,
	}

	err := c.RawQuery(RawQueryFightEvents, variables, &Query)
	if err != nil {
		log.Println(err)

		return nil
	}

	var events []FFLogsEvent

	err = json.Unmarshal(Query.Data.ReportData.Report.Events.Data, &events)
	if err != nil {
		log.Println(err)

		return nil
	}

	ret = append(ret, events...)

	for Query.Data.ReportData.Report.Events.NextPageTimestamp != 0 {
		events = []FFLogsEvent{}
		variables["startTime"] = Query.Data.ReportData.Report.Events.NextPageTimestamp
		Query.Data.ReportData.Report.Events.NextPageTimestamp = 0

		err = c.RawQuery(RawQueryFightEvents, variables, &Query)
		if err != nil {
			log.Println(err)

			return nil
		}

		err = json.Unmarshal(Query.Data.ReportData.Report.Events.Data, &events)
		if err != nil {
			log.Println(err)

			return nil
		}

		ret = append(ret, events...)
	}

	return ret
}

type Ability struct {
	Guid        int64  `json:"guid"`
	Name        string `json:"name"`
	AbilityIcon string `json:"abilityIcon"`
}

func (a Ability) ToBuff() *model.Buff {
	return &model.Buff{
		ID:     a.Guid,
		Name:   a.Name,
		Icon:   a.AbilityIcon,
		Stacks: 1,
	}
}

func (a Ability) ToSkill(duration int64) model.Skill {
	return model.Skill{
		ID:     a.Guid,
		Name:   a.Name,
		Icon:   a.AbilityIcon,
		Cast:   duration,
		Recast: 0,
	}
}

func getFFLogsToken(clientID, clientSecret string) (*Credentials, error) {
	data := []byte(`grant_type=client_credentials`)

	req, err := http.NewRequest(http.MethodPost, "https://www.fflogs.com/oauth/token", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get credentials: %s", resp.Status)
	}

	var creds Credentials
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return nil, err
	}

	return &creds, nil
}
