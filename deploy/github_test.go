package deploy

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/btschwartz12/autodeploy/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func getTestEvent(repo string) *model.PushEvent {
	parts := strings.Split(repo, "/")

	return &model.PushEvent{
		Ref:   "refs/heads/main",
		Owner: parts[0],
		Repo:  parts[1],
	}
}

func getTestService() *model.Service {
	return &model.Service{
		HealthcheckURL: "https://example.com",
	}
}

func TestCreateDeploymentPending(t *testing.T) {
	token, err := os.ReadFile("ghtoken")
	assert.NoError(t, err)
	repoB, err := os.ReadFile("ghrepo")
	assert.NoError(t, err)
	repo := string(repoB)

	deployer := New(zap.NewNop().Sugar(), string(token))
	assert.NoError(t, err)

	id, err := deployer.createDeployment(context.Background(), getTestEvent(repo))
	assert.NoError(t, err)
	assert.NotZero(t, id)

	err = deployer.createDeploymentStatus(
		context.Background(),
		id,
		getTestService(),
		getTestEvent(repo),
		StatePending,
	)
	assert.NoError(t, err)
}

func TestCreateDeploymentSuccess(t *testing.T) {
	token, err := os.ReadFile("ghtoken")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	repoB, err := os.ReadFile("ghrepo")
	assert.NoError(t, err)
	repo := string(repoB)

	deployer := New(zap.NewNop().Sugar(), string(token))
	assert.NoError(t, err)

	id, err := deployer.createDeployment(context.Background(), getTestEvent(repo))
	assert.NoError(t, err)
	assert.NotZero(t, id)

	err = deployer.createDeploymentStatus(
		context.Background(),
		id,
		getTestService(),
		getTestEvent(repo),
		StatePending,
	)
	assert.NoError(t, err)

	err = deployer.createDeploymentStatus(
		context.Background(),
		id,
		getTestService(),
		getTestEvent(repo),
		StateSuccess,
	)
	assert.NoError(t, err)
}

func TestCreateDeploymentFailure(t *testing.T) {
	token, err := os.ReadFile("ghtoken")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	repoB, err := os.ReadFile("ghrepo")
	assert.NoError(t, err)
	repo := string(repoB)

	deployer := New(zap.NewNop().Sugar(), string(token))
	assert.NoError(t, err)

	id, err := deployer.createDeployment(context.Background(), getTestEvent(repo))
	assert.NoError(t, err)
	assert.NotZero(t, id)

	err = deployer.createDeploymentStatus(
		context.Background(),
		id,
		getTestService(),
		getTestEvent(repo),
		StatePending,
	)
	assert.NoError(t, err)

	err = deployer.createDeploymentStatus(
		context.Background(),
		id,
		getTestService(),
		getTestEvent(repo),
		StateFailure,
	)
	assert.NoError(t, err)
}
