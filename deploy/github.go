package deploy

import (
	"context"
	"fmt"
	"net/http"

	"github.com/btschwartz12/autodeploy/model"
	"github.com/google/go-github/v68/github"
)

type State string

const (
	StatePending State = "pending"
	StateSuccess State = "success"
	StateFailure State = "failure"
)

func (d *Deployer) createDeployment(
	ctx context.Context,
	event *model.PushEvent,
) (int64, error) {
	deployment, resp, err := d.client.CreateDeployment(
		ctx,
		event.Owner,
		event.Repo,
		&github.DeploymentRequest{
			Ref: &event.Ref,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create deployment: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("unexpected return code when creating deployment: %d", resp.StatusCode)
	}
	if deployment.Ref == nil {
		return 0, fmt.Errorf("ref not set in deployment")
	}
	if *deployment.Ref != event.Ref {
		return 0, fmt.Errorf("unexpected ref in deployment: %s", *deployment.Ref)
	}
	return deployment.GetID(), nil
}

func (d *Deployer) createDeploymentStatus(
	ctx context.Context,
	deploymentID int64,
	service *model.Service,
	event *model.PushEvent,
	state State,
) error {
	stateStr := string(state)
	status, resp, err := d.client.CreateDeploymentStatus(
		ctx,
		event.Owner,
		event.Repo,
		deploymentID,
		&github.DeploymentStatusRequest{
			State:          &stateStr,
			EnvironmentURL: &service.HealthcheckURL,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to create deployment status: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected return code when creating deployment status: %d", resp.StatusCode)
	}
	if status.State == nil {
		return fmt.Errorf("state not set in deployment status")
	}
	if *status.State != stateStr {
		return fmt.Errorf("unexpected state in deployment status: %s", *status.State)
	}
	return nil
}
