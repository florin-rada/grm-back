package config

import "os"

type Config struct {
	AuthConfig *AuthConfig `json:"auth_config"`
}

type AuthConfig struct {
	ApiURL             string `json:"api_url"`
	GitlabAPIKey       string `json:"gitlab_api_key"`
	GitlabAPISecretKey string `json:"gitlab_api_secret_key"`
	GitlabAuthURL      string `json:"gitlab_auth_url"`
	GitlabTokenURL     string `json:"gitlab_token_url"`
	RedirectURL        string `json:"redirect_url"`
}

func LoadConfig() *Config {
	cfg := &Config{}

	cfg.AuthConfig = &AuthConfig{
		ApiURL:             os.Getenv("GITLAB_API_URL"),
		GitlabAPIKey:       os.Getenv("GITLAB_API_KEY"),
		GitlabAPISecretKey: os.Getenv("GITLAB_SECRET_KEY"),
		GitlabAuthURL:      os.Getenv("GITLAB_AUTH_URL"),
		GitlabTokenURL:     os.Getenv("GITLAB_TOKEN_URL"),
		RedirectURL:        os.Getenv("GITLAB_REDIRECT_URL"),
	}
	if cfg.AuthConfig.ApiURL == "" {
		panic("GITLAB_API_URL environment variable is not set")
	}
	if cfg.AuthConfig.GitlabAPIKey == "" {
		panic("GITLAB_API_KEY environment variable is not set")
	}
	if cfg.AuthConfig.GitlabAPISecretKey == "" {
		panic("GITLAB_SECRET_KEY environment variable is not set")
	}
	return cfg
}

func init() {

}
