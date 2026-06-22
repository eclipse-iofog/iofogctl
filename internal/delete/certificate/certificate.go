package deletecertificate

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/execute"
	clientutil "github.com/eclipse-iofog/iofogctl/internal/util/client"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
)

type Executor struct {
	namespace string
	name      string
}

func NewExecutor(namespace, name string) (execute.Executor, error) {
	exe := &Executor{
		namespace: namespace,
		name:      name,
	}

	return exe, nil
}

// GetName returns application name
func (exe *Executor) GetName() string {
	return exe.name
}

// Execute deletes application by deleting its associated application
func (exe *Executor) Execute() error {
	util.SpinStart("Deleting Certificate")
	// Init remote resources
	clt, err := clientutil.NewControllerClient(exe.namespace)
	if err != nil {
		return err
	}

	if certificate, err := clt.GetCertificate(exe.name); err == nil {
		if certificate.IsCA {
			util.SpinHandlePrompt()
			fmt.Println("Are you sure you want to delete the CA certificate? This might affect all certificates issued by this CA. (y/n)")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading response: %w", err)
			}
			response = strings.TrimSpace(response)
			if strings.ToLower(response) != "y" {
				return util.NewInputError("CA certificate not deleted")
			} else {
				util.SpinHandlePromptComplete()
				return clt.DeleteCA(exe.name)
			}
		} else {
			return clt.DeleteCertificate(exe.name)
		}
	}

	return util.NewNotFoundError("Certificate not found")
}
