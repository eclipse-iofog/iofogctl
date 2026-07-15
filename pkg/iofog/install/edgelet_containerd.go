package install

import (
	"fmt"
	"strings"
	"time"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const (
	edgeletContainerdReadyTimeoutSec = 150
	edgeletContainerdPollIntervalSec = 10
)

func edgeletContainerdReadyCheckShell() string {
	return `systemctl is-active --quiet edgelet-containerd 2>/dev/null && ` +
		`{ _sock=/run/edgelet/containerd.sock; [ -S "$_sock" ] || _sock=/var/run/edgelet/containerd.sock; ` +
		`[ -S "$_sock" ] || exit 1; _ctr=""; ` +
		`if [ -x /var/lib/edgelet/data/current/bin/ctr ]; then _ctr=/var/lib/edgelet/data/current/bin/ctr; ` +
		`elif command -v ctr >/dev/null 2>&1; then _ctr=ctr; fi; ` +
		`[ -z "$_ctr" ] || "$_ctr" --address "$_sock" version >/dev/null 2>&1; }`
}

func restartEdgeletContainerdNoBlockShell() string {
	return "sudo systemctl restart --no-block edgelet-containerd"
}

func markContainerdRestartedShell(stageDir string) string {
	dir := strings.TrimSpace(stageDir)
	if dir == "" {
		dir = EdgeletScriptStageDir
	}
	return fmt.Sprintf("sudo mkdir -p %s && echo 1 | sudo tee %s/.containerd-restarted >/dev/null", dir, dir)
}

func waitForEdgeletContainerdReady(check func() (bool, error)) error {
	util.PrintInfo(fmt.Sprintf(
		"Waiting for edgelet-containerd (up to %ds)",
		edgeletContainerdReadyTimeoutSec,
	))
	deadline := time.Now().Add(edgeletContainerdReadyTimeoutSec * time.Second)
	for elapsed := 0; time.Now().Before(deadline); elapsed += edgeletContainerdPollIntervalSec {
		ready, err := check()
		if err != nil {
			return err
		}
		if ready {
			util.PrintInfo("edgelet-containerd is ready")
			return nil
		}
		util.PrintInfo(fmt.Sprintf(
			"waiting for edgelet-containerd (%ds / %ds)",
			elapsed,
			edgeletContainerdReadyTimeoutSec,
		))
		time.Sleep(edgeletContainerdPollIntervalSec * time.Second)
	}
	return fmt.Errorf("edgelet-containerd not ready after %ds", edgeletContainerdReadyTimeoutSec)
}
