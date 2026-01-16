package jira

type IntegrationConfig struct {
	ProjectUrl  string `json:"project_url"`
	ApiUsername string `json:"api_username"`
	ApiToken    string `json:"api_token"`
}
