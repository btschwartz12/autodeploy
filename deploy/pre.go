package deploy

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/http"

	"github.com/btschwartz12/autodeploy/model"
)

func (d *Deployer) pre(ctx context.Context, service *model.Service, event *model.PushEvent) error {
	err := d.pull(ctx, service, event)
	if err != nil {
		return fmt.Errorf("failed to pull: %w", err)
	}
	d.logger.Infow("pulled", "service", service.Name)

	err = d.build(ctx, service)
	if err != nil {
		return fmt.Errorf("failed to build: %w", err)
	}
	d.logger.Infow("built", "service", service.Name)
	return nil
}

func (d *Deployer) pull(ctx context.Context, service *model.Service, event *model.PushEvent) error {
	repo, err := git.PlainOpen(service.Path)
	if err != nil {
		return fmt.Errorf("failed to open git repo: %w", err)
	}
	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}
	if head.Hash().String() != event.BeforeSha {
		return fmt.Errorf("Latest local commit (%s) does not match before_sha (%s)", head.Hash().String(), event.BeforeSha)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("failed to get worktree: %w", err)
	}
	status, err := worktree.Status()
	if err != nil {
		return fmt.Errorf("failed to get worktree status: %w", err)
	}
	if !status.IsClean() {
		return fmt.Errorf("worktree is not clean")
	}
	err = worktree.PullContext(ctx, &git.PullOptions{
		Force: true,
		Auth: &http.BasicAuth{
			Username: "can-be-anything",
			Password: d.ghToken,
		},
		RemoteName: "autodeploy",
	})
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	}
	if errors.Is(err, git.ErrNonFastForwardUpdate) {
		d.logger.Infow("non-fast-forward update detected, resetting", "service", service.Name)
		newHead := plumbing.NewHash(event.AfterSha)
		err = worktree.Reset(&git.ResetOptions{
			Mode:   git.HardReset,
			Commit: newHead,
		})
		if err != nil {
			return fmt.Errorf("failed to reset worktree: %w", err)
		}
		err = worktree.PullContext(ctx, &git.PullOptions{
			Force: true,
			Auth: &http.BasicAuth{
				Username: "can-be-anything",
				Password: d.ghToken,
			},
			RemoteName: "autodeploy",
		})
		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
			return fmt.Errorf("failed to pull: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to pull: %w", err)
	}
	return nil
}

func (d *Deployer) build(ctx context.Context, service *model.Service) error {
	if service.HasBuildCommand() {
		d.logger.Infow("running build command", "service", service.Name, "command", service.BuildCommand)
		err := runCommand(ctx, service, true, "sh", "-c", service.BuildCommand)
		if err != nil {
			return fmt.Errorf("failed to run build command: %w", err)
		}
	} else {
		d.logger.Infow("no build command specified", "service", service.Name)
	}
	if service.ComposeService {
		d.logger.Infow("building docker compose service", "service", service.Name)
		err := runCommand(ctx, service, true, "docker", "compose", "build")
		if err != nil {
			return fmt.Errorf("failed to run docker compose build: %w", err)
		}
	}
	return nil
}
