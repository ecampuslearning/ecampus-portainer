package options

import "net/http"

type SearchRepoOptions struct {
	Repo     string       `example:"https://charts.gitlab.io/"`
	Client   *http.Client `example:"&http.Client{Timeout: time.Second * 10}"`
	Chart    string       `example:"my-chart"`
	UseCache bool         `example:"false"`
}
