package fflogs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Xinrea/ffreplay/internal/data/markers"
	"github.com/Xinrea/ffreplay/internal/model"
	"github.com/Xinrea/ffreplay/util"
)

const ENDPOINT_CLIENT = "https://cn.fflogs.com/api/v2/client"

const ENDPOINT_USER = "https://cn.fflogs.com/api/v2/user"

type FFLogsClient struct {
	clientID     string
	client       *http.Client
	isAuthorized bool
}

func NewFFLogsClient(authCode, clientID, clientSecret string) *FFLogsClient {
	creds, err := getFFLogsToken(authCode, clientID, clientSecret)
	if err != nil {
		log.Panic(err)
	}

	// create http client with bearer token
	httpClient := &http.Client{
		Transport: &BearerAuthTransport{
			Token: creds.AccessToken,
		},
	}

	return &FFLogsClient{
		clientID:     clientID,
		client:       httpClient,
		isAuthorized: creds.IsAuthorized,
	}
}

func (c *FFLogsClient) Endpoint() string {
	if c.isAuthorized {
		return ENDPOINT_USER
	}

	return ENDPOINT_CLIENT
}

// Have to do this, as fflogs graphql not providing ways to query worldmarkers.
func (c *FFLogsClient) QueryWorldMarkers(code string, fight int) []markers.WorldMarker {
	if util.IsWasm() {
		return markers.QueryWorldMarkersFromApi(code, fight)
	}

	return markers.QueryWorldMarkers(code, fight)
}

func (c *FFLogsClient) RawQuery(query string, variables map[string]any, result any) {
	for k, v := range variables {
		query = strings.ReplaceAll(query, "$"+k, fmt.Sprintf("%v", v))
	}

	requestBody, err := json.Marshal(map[string]string{
		"query": query,
	})
	if err != nil {
		log.Panic(err)
	}

	resp, err := c.client.Post(c.Endpoint(), "application/json", bytes.NewReader(requestBody))
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Panic(err)
	}
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

	c.RawQuery(`
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

	c.RawQuery(`
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

	return Query.Data.ReportData.Report.MasterData.Actors
}

func findFightIndex(fights []ReportFight, fight int) int {
	if fight == -1 {
		return len(fights) - 1
	}

	for i := range fights {
		if fights[i].ID == fight {
			return i
		}
	}

	return -1
}

func (c *FFLogsClient) QueryReportFight(reportCode string, fight int) ReportFight {
	var Query struct {
		Errors []struct {
			Message string
		}
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

	c.RawQuery(`
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

	if len(Query.Errors) > 0 {
		if !c.isAuthorized && strings.Contains(Query.Errors[0].Message, "permission") {
			state := util.GenerateNonce()
			util.UpdateLocalStorage(state, fmt.Sprintf("%s:%d", reportCode, fight))
			// oauth to get token
			util.Redirect(
				fmt.Sprintf(
					"https://www.fflogs.com/oauth/authorize?client_id=%s&redirect_uri=%s&state=%s&response_type=code",
					c.clientID,
					util.CurrentOrigin(),
					state,
				),
			)

			os.Exit(0)
		}

		log.Panic(Query.Errors)
	}

	if len(Query.Data.ReportData.Report.Fights) == 0 {
		log.Panic("No fight found")
	}

	index := findFightIndex(Query.Data.ReportData.Report.Fights, fight)
	if index == -1 {
		log.Panic("Invalid fight id")
	}

	return Query.Data.ReportData.Report.Fights[index]
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

	c.RawQuery(`
			query {
				reportData {
					report(code: "$code") {
						playerDetails(fightIDs: $fightIDs)
					}
				}
			}
		`, variables, &Query)

	var players struct {
		Data struct {
			PlayerDetails PlayerDetails `json:"playerDetails"`
		}
	}

	err := json.Unmarshal(Query.Data.ReportData.Report.PlayerDetails, &players)
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

func (c *FFLogsClient) QueryFightEvents(query string, reportCode string, fight ReportFight) (ret []FFLogsEvent) {
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

	c.RawQuery(query, variables, &Query)

	var events []FFLogsEvent

	err := json.Unmarshal(Query.Data.ReportData.Report.Events.Data, &events)
	if err != nil {
		log.Println(err)

		return nil
	}

	ret = append(ret, events...)

	for Query.Data.ReportData.Report.Events.NextPageTimestamp != 0 {
		events = []FFLogsEvent{}
		variables["startTime"] = Query.Data.ReportData.Report.Events.NextPageTimestamp
		Query.Data.ReportData.Report.Events.NextPageTimestamp = 0

		c.RawQuery(query, variables, &Query)

		err = json.Unmarshal(Query.Data.ReportData.Report.Events.Data, &events)
		if err != nil {
			log.Println(err)

			return nil
		}

		ret = append(ret, events...)
	}

	return ret
}

const RawQueryDamageTakenEvents = `
query {
	reportData {
		report(code: "$code") {
			events(
				fightIDs: $fightIDs,
				startTime: $startTime,
				endTime: $endTime,
				dataType: DamageTaken,
				filterExpression:"type=\"calculateddamage\"",
				limit: 10000,
			) {
				data
				nextPageTimestamp
			}
		}
	}
}
`

type Ability struct {
	Type        int    `json:"type"`
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

func getFFLogsToken(authCode, clientID, clientSecret string) (*Credentials, error) {
	tokenString := util.GetLocalStorage("access_token")
	if tokenString != "" && authCode == "" {
		var creds Credentials

		err := json.Unmarshal([]byte(tokenString), &creds)
		if err != nil {
			log.Println("Failed to unmarshal access token", err)
			util.RemoveLocalStorage("access_token")

			return getFFLogsToken(authCode, clientID, clientSecret)
		}

		// check if expired
		if time.Now().Unix() > creds.ExpiresAt {
			log.Println("Access token expired, removing from local storage")
			util.RemoveLocalStorage("access_token")

			return getFFLogsToken(authCode, clientID, clientSecret)
		}

		log.Println("Using cached token, expires at", time.Unix(creds.ExpiresAt, 0))

		return &creds, nil
	}

	values := url.Values{}
	if authCode != "" {
		values.Set("grant_type", "authorization_code")
		values.Set("code", authCode)
		values.Set("redirect_uri", util.CurrentOrigin())
	} else {
		values.Set("grant_type", "client_credentials")
	}

	req, err := http.NewRequest(http.MethodPost, "https://www.fflogs.com/oauth/token", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// format body to json
		var bodyMap map[string]interface{}

		err = json.Unmarshal(body, &bodyMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse body: %w", err)
		}

		bodyJson, _ := json.MarshalIndent(bodyMap, "", "  ")

		return nil, fmt.Errorf("failed to get credentials: %s\n%s", resp.Status, string(bodyJson))
	}

	var creds Credentials
	if err := json.NewDecoder(resp.Body).Decode(&creds); err != nil {
		return nil, err
	}

	creds.IsAuthorized = authCode != ""
	creds.ExpiresAt = time.Now().Unix() + int64(creds.ExpiresIn)/1000 - 60

	log.Println("Update access token, expires at", time.Unix(creds.ExpiresAt, 0))
	credsJson, _ := json.Marshal(creds)

	util.UpdateLocalStorage("access_token", string(credsJson))

	return &creds, nil
}
