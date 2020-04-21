package scripts

import (
	"encoding/json"
	"errors"
	"github.com/mdg-iitr/Codephile/models/types"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type HackerrankGraphPoint struct {
	ContestName string
	Date        string
	Rating      float64
}

type Contests struct {
	Data  []HackerrankContest `json:"models"`
	Count int                 `json:"total"`
}

type HackerrankContest struct {
	ContestName string `json:"name"`
	Rated       bool   `json:"rated"`
	EpochStart  int64  `json:"epoch_starttime"`
	EpochEnd    int64  `json:"epoch_endtime"`
	Archived    bool   `json:"archived"`
}

func GetRequest(path string) []byte {
	client := http.Client{Timeout: time.Second * 10}
	resp, err := client.Get(path)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer resp.Body.Close() // nolint: errcheck
	byteValue, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	return byteValue
}

func GetHackerrankProfileInfo(handle string) types.ProfileInfo {
	path := "https://www.hackerrank.com/rest/contests/master/hackers/" + handle + "/profile";
	byteValue := GetRequest(path)
	if byteValue == nil {
		log.Println(errors.New("GetRequest failed. Please check connection status"))
		return types.ProfileInfo{}
	}
	var JsonInterFace interface{}
	err := json.Unmarshal(byteValue, &JsonInterFace)
	if err != nil {
		log.Println(err.Error())
		return types.ProfileInfo{}
	}
	Profile := JsonInterFace.(map[string]interface{})["model"].(map[string]interface{})
	Name := Profile["name"].(string)
	// Date := Profile["created_at"].(string)
	UserName := Profile["username"].(string)
	School := Profile["school"].(string)
	return types.ProfileInfo{Name: Name, UserName: UserName, School: School}
}

func GetHackerrankSubmissions(handle string, after time.Time) []types.Submission {
	path := "https://www.hackerrank.com/rest/hackers/" + handle + "/recent_challenges?limit=1000&response_version=v1"
	byteValue := GetRequest(path)
	var data types.HackerrankSubmisson
	err := json.Unmarshal(byteValue, &data)
	submissions := data.Models
	if err != nil {
		log.Println(err.Error())
		return nil
	}
	var oldestSubIndex int;
	if after.IsZero() {
		oldestSubIndex = len(submissions)
	} else {
		for i, sub := range submissions {
			if sub.CreationDate.Equal(after) || sub.CreationDate.Before(after) {
				oldestSubIndex = i
				break
			}
		}
	}
	submissions = submissions[0:oldestSubIndex]
	for i := 0; i < len(submissions); i++ {
		submissions[i].URL = "https://www.hackerrank.com" + submissions[i].URL
	}
	return submissions
}

func GetHackerrankContests() Contests {
	path := "https://www.hackerrank.com/rest/contests/upcoming?offset=0&limit=20&contest_slug=active"
	byteValue := GetRequest(path)
	var ContestsArray Contests
	err := json.Unmarshal(byteValue, &ContestsArray)
	if err != nil {
		log.Println(err.Error())
	}
	return ContestsArray
}

func GetHackerrankGraphData(handle string) []HackerrankGraphPoint {
	path := "https://www.hackerrank.com/rest/hackers/" + handle + "/rating_histories_elo"
	byteValue := GetRequest(path)
	var JsonInterFace interface{}
	err := json.Unmarshal(byteValue, &JsonInterFace)
	if err != nil {
		log.Println(err.Error())
	}
	m := JsonInterFace.(map[string]interface{})

	models := m["models"].([]interface{})
	events := models[0].(map[string]interface{})["events"].([]interface{})
	var Graph []HackerrankGraphPoint
	for i := 0; i < len(events); i++ {
		contest := events[i].(map[string]interface{})
		name := contest["contest_name"].(string)
		date := contest["date"].(string)
		rating := contest["rating"].(float64)
		Graph = append(Graph, HackerrankGraphPoint{name, date, rating})
	}
	return Graph
}
func CheckHackerrankHandle(handle string) bool {
	resp, err := http.Get("https://www.hackerrank.com/rest/contests/master/hackers/" + handle + "/profile")
	if err != nil {
		log.Println(err.Error())
		return false
	}
	return resp.StatusCode != http.StatusNotFound
}
