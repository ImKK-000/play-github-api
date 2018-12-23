package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Config struct {
	URL string `json:"url"`
}

type AuthConfig struct {
	ClientID, ClientSecret string
}

type GithubUser struct {
	ID uint `json:"id"`
}

type GithubCommit struct {
	CommitInformation CommitInformation `json:"commit"`
}

type CommitInformation struct {
	Message   string    `json:"message"`
	Committer Committer `json:"committer"`
}

type Committer struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Date  string `json:"date"`
}

var responseStore [][]byte
var auth_token string

func APIWalk(config Config, githubUser GithubUser, url string) error {
	response, _ := http.Get(url)
	responseBody, _ := ioutil.ReadAll(response.Body)
	responseStore = append(responseStore, responseBody)

	headerLinkInformation := response.Header.Get("Link")
	splitComma := strings.Split(headerLinkInformation, ",")
	if len(splitComma) <= 1 {
		return errors.New("Response header[Link]: length <= 1")
	}

	isNextLinkAtIndexZero := strings.HasSuffix(splitComma[0], `rel="next"`)
	isNextLinkAtIndexOne := strings.HasSuffix(splitComma[1], `rel="next"`)
	if !(isNextLinkAtIndexZero || isNextLinkAtIndexOne) {
		return errors.New("Response header[Link]: Not found 'rel=next'")
	}
	rawNextLink := splitComma[1]
	if isNextLinkAtIndexZero {
		rawNextLink = splitComma[0]
	}
	nextLink := strings.TrimSpace(rawNextLink)
	nextLink = strings.TrimLeft(nextLink, "<")
	nextLink = strings.TrimRight(nextLink, `>; rel="next"`)
	return APIWalk(config, githubUser, nextLink)
}

func ConvertToUnixTimestamp(date string) string {
	dateTime, _ := time.Parse(time.RFC3339, date)
	dateFormat := dateTime.Format("02/01/2006")
	timeFormat := dateTime.Format("15:04:05")
	return fmt.Sprintf("%s %s", dateFormat, timeFormat)
}

func main() {
	rawConfig, _ := ioutil.ReadFile("config/api.json")
	rawAuthConfig, _ := ioutil.ReadFile("config/auth.json")

	var config Config
	var authConfig AuthConfig
	json.Unmarshal(rawConfig, &config)
	json.Unmarshal(rawAuthConfig, authConfig)
	auth_token = fmt.Sprintf("&client_id=%s&client_secret=%s", authConfig.ClientID, authConfig.ClientSecret)

	response, _ := http.Get(config.URL + "/users/imkk-000")
	responseBody, _ := ioutil.ReadAll(response.Body)

	var githubUser GithubUser
	json.Unmarshal(responseBody, &githubUser)
	firstPageURL := config.URL + "/repos/ImKK-000/workflow-designer/commits?page=1&sha=develop%402.x.x" + auth_token
	APIWalk(config, githubUser, firstPageURL)

	for _, response := range responseStore {
		var gitCommits []GithubCommit
		json.Unmarshal(response, &gitCommits)
		for _, commit := range gitCommits {
			fmt.Println(ConvertToUnixTimestamp(commit.CommitInformation.Committer.Date))
		}
	}
}
