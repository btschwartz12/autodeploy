package deploy

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
	"go.uber.org/zap"

	"github.com/btschwartz12/autodeploy/model"
	"github.com/btschwartz12/autodeploy/slack"
)

type Deployer struct {
	logger  *zap.SugaredLogger
	client  *github.RepositoriesService
	ghToken string
	slack   *slack.SlackClient
}

func New(logger *zap.SugaredLogger, githubToken string) *Deployer {
	ghClient := github.NewClient(nil).WithAuthToken(githubToken)
	return &Deployer{
		logger:  logger,
		client:  ghClient.Repositories,
		ghToken: githubToken,
	}
}

func (d *Deployer) Deploy(ctx context.Context, service *model.Service, event *model.PushEvent) error {
	// make deployment
	d.logger.Infow("beginning deployment", "service", service.Name)
	deploymentID, err := d.notifyBegin(ctx, service, event)
	if err != nil {
		return fmt.Errorf("failed to notify: %w", err)
	}
	// pre-activation
	d.logger.Infow("pre-activation", "service", service.Name)
	err = d.pre(ctx, service, event)
	if err != nil {
		notifyErr := d.notifyFinish(ctx, deploymentID, service, event, StateFailure)
		if notifyErr != nil {
			d.logger.Errorw("failed to notify failure", "error", notifyErr)
		}
		return fmt.Errorf("pre-activation failed: %w", err)
	}
	// activation
	d.logger.Infow("activation", "service", service.Name)
	err = d.activate(ctx, service)
	if err != nil {
		notifyErr := d.notifyFinish(ctx, deploymentID, service, event, StateFailure)
		if notifyErr != nil {
			d.logger.Errorw("failed to notify failure", "error", notifyErr)
		}
		return fmt.Errorf("activation failed: %w", err)
	}
	// post-activation
	d.logger.Infow("post-activation", "service", service.Name)
	err = d.post(ctx, service)
	if err != nil {
		notifyErr := d.notifyFinish(ctx, deploymentID, service, event, StateFailure)
		if notifyErr != nil {
			d.logger.Errorw("failed to notify failure", "error", notifyErr)
		}
		return fmt.Errorf("post-activation failed: %w", err)
	}
	// success
	notifyErr := d.notifyFinish(ctx, deploymentID, service, event, StateSuccess)
	if notifyErr != nil {
		d.logger.Errorw("failed to notify success", "error", notifyErr)
	}
	return nil
}

func (d *Deployer) notifyBegin(
	ctx context.Context,
	service *model.Service,
	event *model.PushEvent,
) (int64, error) {
	deploymentID, err := d.createDeployment(ctx, event)
	if err != nil {
		return 0, fmt.Errorf("failed to create deployment: %w", err)
	}
	err = d.createDeploymentStatus(ctx, deploymentID, service, event, StatePending)
	if err != nil {
		return 0, fmt.Errorf("failed to create deployment status: %w", err)
	}
	d.logger.Infow("created pending deployment", "deployment_id", deploymentID)
	return deploymentID, nil
}

func (d *Deployer) notifyFinish(
	ctx context.Context,
	deploymentID int64,
	service *model.Service,
	event *model.PushEvent,
	state State,
) error {
	err := d.createDeploymentStatus(ctx, deploymentID, service, event, state)
	if err != nil {
		return fmt.Errorf("failed to create deployment status: %w", err)
	}
	d.logger.Infow("created deployment status", "deployment_id", deploymentID, "state", state)
	return nil
}
