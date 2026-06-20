package cmd

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/eclipse-iofog/iofogctl/internal/config"
	"github.com/eclipse-iofog/iofogctl/pkg/iofog"
	"github.com/eclipse-iofog/iofogctl/pkg/util"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

func newViewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Open ECN Viewer",
		Run: func(cmd *cobra.Command, args []string) {
			// Get Control Plane
			namespace, err := cmd.Flags().GetString("namespace")
			util.Check(err)
			ns, err := config.GetNamespace(namespace)
			util.Check(err)
			if len(ns.GetControllers()) == 0 {
				util.PrintError("You must deploy a Control Plane to a namespace to see an ECN Viewer")
				os.Exit(1)
			}
			cp, err := ns.GetControlPlane()
			util.Check(err)
			cpEndpoint, err := cp.GetEndpoint()
			if err != nil {
				util.PrintError("Failed to get Control Plane endpoint: " + err.Error())
				os.Exit(1)
			}
			ctrl := ns.GetControllers()[0]
			var endpoint string
			if cpEndpoint != "" {
				endpoint = cpEndpoint
			} else {
				endpoint = ctrl.GetEndpoint()
			}
			URL, err := url.Parse(endpoint)
			if err != nil || URL.Host == "" {
				URL, err = url.Parse("//" + ctrl.GetEndpoint()) // Try to see if controllerEndpoint is an IP, in which case it needs to be pefixed by //
			}
			util.Check(err)
			if URL.Scheme == "" {
				URL.Scheme = "http"
			}
			host := ""
			if strings.Contains(URL.Host, ":") {
				host, _, err = net.SplitHostPort(URL.Host)
				util.Check(err)
			} else {
				host = URL.Host
			}
			if util.IsLocalHost(host) {
				host += ":" + iofog.ControllerHostECNViewerPortString
			}
			URL.Host = host
			ecnViewer := URL.String()
			if err := browser.OpenURL(ecnViewer); err != nil {
				util.PrintInfo("To see the ECN Viewer, open your browser and go to:\n")
				util.PrintInfo(fmt.Sprintf("%s\n", URL))
			}
		},
	}
	return cmd
}
