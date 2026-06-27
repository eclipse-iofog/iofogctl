package deployofflineimage

import (
	"fmt"
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

const remoteOfflineDir = "/tmp/iofogctl-offline"

func transferArtifact(plan agentPlan, artifact *imageArtifact) error {
	ssh, err := util.NewSecureShellClient(plan.agent.SSH.User, plan.agent.Host, plan.agent.SSH.KeyFile)
	if err != nil {
		return err
	}
	ssh.SetPort(plan.agent.SSH.Port)
	if err := ssh.Connect(); err != nil {
		return err
	}
	defer util.Log(ssh.Disconnect)

	agentDir := util.JoinAgentPath(remoteOfflineDir, sanitizeSegment(plan.agent.Name))
	if err := ssh.CreateFolder(agentDir); err != nil {
		return err
	}

	file, err := util.OpenValidatedFile(artifact.path)
	if err != nil {
		return err
	}
	defer util.IgnoreClose(file)
	info, err := file.Stat()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s-%s.tar.gz", sanitizeSegment(plan.agent.Name), strings.ReplaceAll(plan.platform, "/", "_"))
	label := fmt.Sprintf("Transferring %s to %s", plan.platform, plan.agent.Name)
	reader := newProgressReader(file, info.Size(), label)
	copyErr := ssh.CopyTo(reader, util.AddTrailingSlash(agentDir), filename, "0600", info.Size())
	reader.Close()
	if copyErr != nil {
		return copyErr
	}
	remotePath := util.JoinAgentPath(agentDir, filename)

	loadCmd := fmt.Sprintf("sudo -S %s load -i %s", plan.engine.command(), remotePath)
	if _, err := ssh.Run(loadCmd); err != nil {
		return err
	}

	if _, err := ssh.Run("sudo rm -f " + remotePath); err != nil {
		return err
	}

	util.PrintInfo(fmt.Sprintf("%s transfer to %s complete", plan.platform, plan.agent.Name))
	return nil
}
