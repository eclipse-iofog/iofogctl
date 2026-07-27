package install

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const (
	edgeletContainerdReadyTimeoutSec = 150
	edgeletContainerdPollIntervalSec = 10
)

func edgeletScriptPath(stageDir, script string) string {
	dir := strings.TrimSpace(stageDir)
	if dir == "" {
		dir = EdgeletScriptStageDir
	}
	return util.JoinAgentPath(dir, script)
}

func runEmbeddedProbeScript(script string) error {
	content, err := loadEdgeletScript(script)
	if err != nil {
		return err
	}
	_, err = util.Exec("", "sh", "-c", content)
	return err
}

func containerdReadyProbe(stageDir string) (bool, error) {
	path := edgeletScriptPath(stageDir, pkg.edgeletScriptProbeContainerdReady)
	if _, err := os.Stat(path); err == nil {
		_, err := util.Exec("", "sh", path)
		return err == nil, nil
	}
	err := runEmbeddedProbeScript(pkg.edgeletScriptProbeContainerdReady)
	return err == nil, nil
}

func restartEdgeletContainerdCommand(stageDir string) string {
	return "sudo " + edgeletScriptPath(stageDir, pkg.edgeletScriptRestartContainerd)
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

func pollContainerdReadyLocal(stageDir string) error {
	return waitForEdgeletContainerdReady(func() (bool, error) {
		return containerdReadyProbe(stageDir)
	})
}
