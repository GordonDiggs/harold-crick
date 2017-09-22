package main

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/peterhellberg/link"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
)

type License struct {
	ShortName string `json:"spdx_id"`
}

type Repo struct {
	Name    string  `json:"full_name"`
	License License `json:"license"`
	Private bool    `json:"private"`
}

type Repos []Repo

func (slice Repos) Len() int {
	return len(slice)
}

func (slice Repos) Less(i, j int) bool {
	return slice[i].Name < slice[j].Name
}

func (slice Repos) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

var apiKey string

func getRepos(url string) Repos {
	client := http.Client{}
	req, reqErr := http.NewRequest(http.MethodGet, url, nil)
	if reqErr != nil {
		log.Fatal(reqErr)
	}
	req.Header.Set("Accept", "application/vnd.github.drax-preview+json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", apiKey))

	res, getErr := client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	repos := make(Repos, 0)
	jsonErr := json.Unmarshal(body, &repos)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	for _, l := range link.ParseResponse(res) {
		if l.Rel == "next" {
			repos = append(repos, getRepos(l.URI)...)
		}
	}

	return repos
}

func listRepos(org string) Repos {
	url := fmt.Sprintf("https://api.github.com/orgs/%s/repos", org)
	return getRepos(url)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./harold-crick <org>")
		os.Exit(1)
	}

	apiKey = os.Getenv("GITHUB_API_KEY")
	if apiKey == "" {
		fmt.Println("You must specify an API key in the environment as `GITHUB_API_KEY`")
		os.Exit(1)
	}

	org := os.Args[1]
	repos := listRepos(org)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Private", "License"})

	sort.Sort(repos)

	for _, repo := range repos {
		row := []string{repo.Name, strconv.FormatBool(repo.Private), repo.License.ShortName}
		table.Append(row)
	}

	table.Render()
}
