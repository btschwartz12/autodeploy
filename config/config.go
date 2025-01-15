package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/btschwartz12/autodeploy/model"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"gopkg.in/yaml.v3"
)

const (
	defaultFlowTimeout = model.Duration(5 * time.Minute)
)

func New(yamlPath string, testFlag bool) (*model.Config, error) {
	configBytes, err := os.ReadFile(yamlPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("config file not found: %s", yamlPath)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	c := &model.Config{}
	if err := yaml.Unmarshal(configBytes, c); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if c.WebhookSecret == "" {
		return nil, fmt.Errorf("webhook_secret must be set")
	}

	if c.WebhookURLSuffix == "" {
		return nil, fmt.Errorf("webhook_url_suffix must be set")
	}

	if c.GithubToken == "" {
		return nil, fmt.Errorf("github_token must be set")
	}

	if c.Hostname == "" {
		return nil, fmt.Errorf("hostname must be set")
	}

	if len(c.Services) == 0 {
		return nil, fmt.Errorf("at least one service must be defined")
	}

	for name, s := range c.Services {
		if err := validate(&s, testFlag); err != nil {
			return nil, fmt.Errorf("service %s: %w", name, err)
		}
		s.Name = name
		s.Hostname = c.Hostname
		c.Services[name] = s
	}

	return c, nil
}

func validate(s *model.Service, testFlag bool) error {
	if s.Repo == "" {
		return fmt.Errorf("repo field must be set")
	}
	if s.Path == "" {
		return fmt.Errorf("path must be set")
	}
	if s.HealthcheckURL == "" {
		return fmt.Errorf("healthcheck_url must be set")
	}
	if s.HasSystemdService() && s.ComposeService {
		return fmt.Errorf("systemd_service and compose_service are mutually exclusive")
	}
	if s.FlowTimeout == 0 {
		s.FlowTimeout = model.Duration(defaultFlowTimeout)
	}
	fileInfo, err := os.Stat(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("path does not exist: %s", s.Path)
		}
		return fmt.Errorf("failed to stat path: %w", err)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", s.Path)
	}
	// make sure there is a .git directory
	gitDir := filepath.Join(s.Path, ".git")
	_, err = os.Stat(gitDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("path is not a git repository: %s", s.Path)
		}
		return fmt.Errorf("failed to stat .git directory: %w", err)
	}

	// sue me
	if testFlag {
		return nil
	}

	// make a remote to push to
	r, err := git.PlainOpen(s.Path)
	if err != nil {
		return fmt.Errorf("failed to open git repository: %w", err)
	}
	_, err = r.CreateRemote(&config.RemoteConfig{
		Name: "autodeploy",
		URLs: []string{fmt.Sprintf("https://github.com/%s", s.Repo)},
	})
	if err != nil && !errors.Is(err, git.ErrRemoteExists) {
		return fmt.Errorf("failed to create remote: %w", err)
	}
	return nil
}
