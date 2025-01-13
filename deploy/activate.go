package deploy

import (
	"context"
	"fmt"

	"github.com/btschwartz12/autodeploy/model"
)

func (d *Deployer) activate(ctx context.Context, service *model.Service) error {
	if service.HasSystemdService() {
		err := runCommand(ctx, service, false, "systemctl", "restart", service.Name)
		if err != nil {
			return fmt.Errorf("failed to restart systemd service: %w", err)
		}
		d.logger.Infow("restarted systemd service", "service", service.Name)
	}
	if service.ComposeService {
		err := runCommand(ctx, service, true, "docker", "compose", "up", "-d")
		if err != nil {
			return fmt.Errorf("failed to start docker-compose service: %w", err)
		}
		d.logger.Infow("started docker-compose service", "service", service.Name)
	}
	return nil
}
