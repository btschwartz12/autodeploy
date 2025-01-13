package deploy

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/btschwartz12/autodeploy/model"
)

func runCommand(ctx context.Context, service *model.Service, forceNoSudo bool, command ...string) error {
	var stdout, stderr bytes.Buffer
	var cmd *exec.Cmd
	if service.NeedsSudo && !forceNoSudo {
		cmd = exec.CommandContext(ctx, "sudo", command...)
	} else {
		cmd = exec.CommandContext(ctx, command[0], command[1:]...)
	}
	cmd.Dir = service.Path
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run command: %w\n%s", err, stderr.String())
	}
	return nil
}
