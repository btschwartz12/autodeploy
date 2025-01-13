package model

import (
	"path/filepath"
)

type Service struct {
	Name             string
	Repo             string   `yaml:"repo"`
	Path             string   `yaml:"path"`
	SystemdService   string   `yaml:"systemd_service"`
	HealthcheckURL   string   `yaml:"healthcheck_url"`
	ComposeService   bool     `yaml:"compose_service"`
	NeedsSudo        bool     `yaml:"needs_sudo"`
	BuildCommand     string   `yaml:"build_command"`
	TriggerWorkflows []string `yaml:"trigger_workflows"`
}

type Config struct {
	GithubToken      string             `yaml:"github_token"`
	WebhookSecret    string             `yaml:"webhook_secret"`
	WebhookURLSuffix string             `yaml:"webhook_url_suffix"`
	Services         map[string]Service `yaml:"services"`
}

func (s *Service) GitDir() string {
	return filepath.Join(s.Path, ".git")
}

func (s *Service) HasSystemdService() bool {
	return s.SystemdService != ""
}

func (s *Service) HasBuildCommand() bool {
	return s.BuildCommand != ""
}

func (c *Config) GetServiceByRepo(repo string) *Service {

	for _, s := range c.Services {
		if s.Repo == repo {
			return &s
		}
	}
	return nil
}
