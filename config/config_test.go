package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/btschwartz12/autodeploy/model"
	"github.com/stretchr/testify/assert"
)

const exampleConfig = `
hostname: your-hostname
webhook_secret: your-webhook-secret
webhook_url_suffix: /postreceive
github_token: your-github-token

services:
  service1:
    repo: "https://github.com/example/repo1"
    path: "/path/to/service1"
    systemd_service: "service1"
    healthcheck_url: "http://localhost:8080/health"
    compose_service: false
    build_command: "make build"
    flow_timeout: 5m

  service2:
    repo: "https://github.com/example/repo2"
    path: "/path/to/service2"
    systemd_service: "service2"
    healthcheck_url: "http://localhost:9090/health"
    compose_service: false
    build_command: "go build ./..."
    trigger_workflows:
      - "deploy-other-thing"
      - "notify"

  service3:
    repo: "https://github.com/example/repo3"
    path: "/path/to/service3"
    healthcheck_url: "http://localhost:3000/health"
    compose_service: true
`

func TestGenerateConfigWithValidPaths(t *testing.T) {
	tmpDir := t.TempDir()

	service1Path := filepath.Join(tmpDir, "service1")
	service2Path := filepath.Join(tmpDir, "service2")
	service3Path := filepath.Join(tmpDir, "service3")

	assert.NoError(t, os.MkdirAll(filepath.Join(service1Path, ".git"), 0755))
	assert.NoError(t, os.MkdirAll(filepath.Join(service2Path, ".git"), 0755))
	assert.NoError(t, os.MkdirAll(filepath.Join(service3Path, ".git"), 0755))

	updatedConfig := exampleConfig
	updatedConfig = strings.ReplaceAll(updatedConfig, "/path/to/service1", service1Path)
	updatedConfig = strings.ReplaceAll(updatedConfig, "/path/to/service2", service2Path)
	updatedConfig = strings.ReplaceAll(updatedConfig, "/path/to/service3", service3Path)

	yamlPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(yamlPath, []byte(updatedConfig), 0644)
	assert.NoError(t, err)
	assert.NoError(t, err)

	config, err := New(yamlPath, true)
	assert.NoError(t, err)

	assert.Equal(t, "your-hostname", config.Hostname)
	assert.Equal(t, "/postreceive", config.WebhookURLSuffix)
	assert.Equal(t, "your-webhook-secret", config.WebhookSecret)
	assert.Equal(t, "your-github-token", config.GithubToken)

	assert.Contains(t, config.Services, "service1")
	assert.Contains(t, config.Services, "service2")
	assert.Contains(t, config.Services, "service3")

	service1 := config.Services["service1"]
	assert.Equal(t, "service1", service1.Name)
	assert.Equal(t, time.Duration(5*time.Minute).String(), service1.FlowTimeout.String())
	assert.Equal(t, "https://github.com/example/repo1", service1.Repo)
	assert.Equal(t, service1Path, service1.Path)
	assert.Equal(t, "service1", service1.SystemdService)
	assert.Equal(t, "http://localhost:8080/health", service1.HealthcheckURL)
	assert.False(t, service1.ComposeService)
	assert.Equal(t, "make build", service1.BuildCommand)

	service2 := config.Services["service2"]
	assert.Equal(t, "service2", service2.Name)
	assert.Equal(t, time.Duration(defaultFlowTimeout).String(), service2.FlowTimeout.String())
	assert.Equal(t, "https://github.com/example/repo2", service2.Repo)
	assert.Equal(t, service2Path, service2.Path)
	assert.Equal(t, "service2", service2.SystemdService)
	assert.Equal(t, "http://localhost:9090/health", service2.HealthcheckURL)
	assert.False(t, service2.ComposeService)
	assert.Equal(t, "go build ./...", service2.BuildCommand)
	assert.ElementsMatch(t, []string{"deploy-other-thing", "notify"}, service2.TriggerWorkflows)

	service3 := config.Services["service3"]
	assert.Equal(t, "service3", service3.Name)
	assert.Equal(t, time.Duration(defaultFlowTimeout).String(), service3.FlowTimeout.String())
	assert.Equal(t, "https://github.com/example/repo3", service3.Repo)
	assert.Equal(t, service3Path, service3.Path)
	assert.Equal(t, "http://localhost:3000/health", service3.HealthcheckURL)
	assert.True(t, service3.ComposeService)
}

func TestServiceValidation(t *testing.T) {
	s := &model.Service{}
	err := validate(s, true)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "repo field must be set")

	s = &model.Service{
		Repo: "ff",
	}
	err = validate(s, true)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "path must be set")

	s = &model.Service{
		Repo: "ff",
		Path: "ff",
	}
	err = validate(s, true)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "healthcheck_url must be set")

	s = &model.Service{
		Repo:           "ff",
		Path:           "ff",
		HealthcheckURL: "ff",
		ComposeService: true,
		SystemdService: "ff",
	}
	err = validate(s, true)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "systemd_service and compose_service are mutually exclusive")
}
