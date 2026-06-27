package exec

import (
	"fmt"

	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

func runExecSession(namespace, resourceLabel string, dial func(*client.Client) (*client.ExecSession, error)) error {
	util.SpinStart(fmt.Sprintf("Connecting Exec Session to %s", resourceLabel))

	clt, err := clientutil.NewControllerClient(namespace)
	if err != nil {
		return err
	}

	util.SpinHandlePrompt()
	session, err := dial(clt)
	if err != nil {
		util.SpinHandlePromptComplete()
		return util.NewError(formatExecError(err))
	}
	util.SpinStop()

	term := newInteractiveTerminal(session)
	if err := term.run(); err != nil {
		return util.NewError(formatExecError(err))
	}

	util.WriteStdout([]byte("\n"))
	util.PrintSuccess(fmt.Sprintf("Successfully closed %s Exec Session", resourceLabel))
	return nil
}

func execDialOptions() *client.DialExecOptions {
	return &client.DialExecOptions{
		OnStatusLine: writeExecStatusLine,
	}
}

func writeExecStatusLine(line string) {
	data := []byte(line)
	if len(data) == 0 || data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	util.WriteStdout(data)
}

func dialExecForMicroservice(clt *client.Client, msvc *client.MicroserviceInfo, isSystem bool) (*client.ExecSession, error) {
	opts := execDialOptions()
	if isSystem {
		return clt.DialSystemMicroserviceExecWithOptions(msvc.UUID, opts)
	}
	return clt.DialMicroserviceExecWithOptions(msvc.UUID, opts)
}

func dialMicroserviceExec(clt *client.Client, fqName string) (*client.ExecSession, error) {
	msvc, isSystem, err := lookupMicroservice(clt, fqName)
	if err != nil {
		return nil, err
	}
	if msvc.Status.Status != "RUNNING" {
		return nil, util.NewError(ErrMsgMicroserviceNotRunning)
	}
	return dialExecForMicroservice(clt, msvc, isSystem)
}
