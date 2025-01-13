package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/btschwartz12/autodeploy/model"
)

const sleepTime = 10 * time.Second

func (d *Deployer) post(ctx context.Context, service *model.Service) error {
	d.logger.Infow("sleeping", "service", service.Name, "duration", sleepTime)
	time.Sleep(sleepTime)
	if service.HasSystemdService() {
		err := runCommand(ctx, service, false, "systemctl", "is-active", "--quiet", service.Name)
		if err != nil {
			return fmt.Errorf("could not get healthy status of systemd service: %w", err)
		}
	}
	if service.ComposeService {
		err := runCommand(ctx, service, true, "sh", "-c", "docker compose ps | grep -q \"Up\" || exit 1")
		if err != nil {
			return fmt.Errorf("could not get healthy status of docker-compose service: %w", err)
		}
	}
	return nil
}
