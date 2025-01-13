package deploy

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/btschwartz12/autodeploy/model"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPull(t *testing.T) {
	token, err := os.ReadFile("ghtoken")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	auth := &http.BasicAuth{
		Username: "abc123", // yes, this can be anything except an empty string
		Password: string(token),
	}

	urlB, err := os.ReadFile("ghrepo")
	assert.NoError(t, err)
	url := fmt.Sprintf("https://github.com/%s", string(urlB))

	// first, clone the repo so we have an 'old' version
	// that will be updated in the test
	oldRepoDir := t.TempDir()
	var out bytes.Buffer
	oldRepo, err := git.PlainClone(oldRepoDir, false, &git.CloneOptions{
		Auth:     auth,
		URL:      url,
		Progress: &out,
	})
	assert.NoError(t, err)
	assert.NotNil(t, oldRepo)
	// create an .env file to make sure it's not deleted
	assert.NoError(t, os.WriteFile(filepath.Join(oldRepoDir, "test.env"), []byte("test"), 0644))
	// set origin to the https url
	_, err = oldRepo.CreateRemote(&config.RemoteConfig{
		Name: "autodeploy",
		URLs: []string{url},
	})
	assert.NoError(t, err)
	w, err := oldRepo.Worktree()
	assert.NoError(t, err)
	status, err := w.Status()
	assert.NoError(t, err)
	assert.True(t, status.IsClean())

	// now, make a repo and make a new commit
	repoDir := t.TempDir()
	out.Reset()
	newRepo, err := git.PlainClone(repoDir, false, &git.CloneOptions{
		Auth:     auth,
		URL:      url,
		Progress: &out,
	})
	assert.NoError(t, err)
	assert.NotNil(t, newRepo)
	// set origin to the https url
	_, err = newRepo.CreateRemote(&config.RemoteConfig{
		Name: "autodeploy",
		URLs: []string{url},
	})
	ref, err := newRepo.Head()
	assert.NoError(t, err)
	beforeSha := ref.Hash().String()
	assert.NotEmpty(t, beforeSha)
	w, err = newRepo.Worktree()
	assert.NoError(t, err)
	// add a file
	assert.NoError(t, os.WriteFile(filepath.Join(repoDir, uuid.New().String()), []byte("test"), 0644))
	_, err = w.Add(".")
	assert.NoError(t, err)
	// commit the file
	_, err = w.Commit("test commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "John Doe",
			Email: "john@doe.org",
			When:  time.Now(),
		},
	})
	assert.NoError(t, err)
	// push the commit
	err = newRepo.Push(&git.PushOptions{
		Auth: auth,
	})
	assert.NoError(t, err)
	// now we can do the thing
	deployer := New(zap.NewNop().Sugar(), string(token))
	err = deployer.pull(
		context.Background(),
		&model.Service{
			Name: "test",
			Repo: "test",
			Path: oldRepoDir,
		},
		&model.PushEvent{
			Ref:       "refs/heads/main",
			BeforeSha: beforeSha,
		},
	)
	assert.NoError(t, err)
	// make sure .env is in the .gitignore of oldRepoDir
	gitignore, err := os.ReadFile(filepath.Join(oldRepoDir, ".gitignore"))
	assert.NoError(t, err)
	assert.Contains(t, string(gitignore), "test.env")
	// make sure the .env file is still there
	_, err = os.Stat(filepath.Join(oldRepoDir, "test.env"))
	assert.NoError(t, err)
}

func TestBuild(t *testing.T) {
	tmpDir := t.TempDir()
	service := &model.Service{
		Name:         "test",
		Path:         tmpDir,
		BuildCommand: "echo 'hello, world!'",
	}
	deployer := New(zap.NewNop().Sugar(), "")
	err := deployer.build(context.Background(), service)
	assert.NoError(t, err)

	exampleDockerfile := "FROM ubuntu:latest"
	assert.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte(exampleDockerfile), 0644))
	exampleComposefile := `
services:
  test:
    build: .
`
	assert.NoError(t, os.WriteFile(filepath.Join(tmpDir, "docker-compose.yml"), []byte(exampleComposefile), 0644))
	service.ComposeService = true
	err = deployer.build(context.Background(), service)
	assert.NoError(t, err)
}
