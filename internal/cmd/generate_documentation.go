package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/eclipse-iofog/iofogctl/pkg/util"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func newGenerateDocumentationCommand(rootCmd *cobra.Command) *cobra.Command {
	// Find home directory.
	home, err := homedir.Dir()
	var docDir string
	util.Check(err)
	cmd := &cobra.Command{
		Use:    "documentation TYPE",
		Hidden: true,
		Short:  ex("Generate %[1]s documentation"),
		Long:   ex("Generate %[1]s documentation as markdown or man page"),
		Example: ex(`%[1]s documentation md
%[1]s documentation man`),
		Args: cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if docDir == "" {
				docDir = home + "/.iofog/docs/"
				err = os.MkdirAll(docDir, util.DirPerm)
				util.Check(err)
			}
			switch t := strings.ToLower(args[0]); t {
			case "md":
				err = os.MkdirAll(docDir, util.DirPerm)
				util.Check(err)
				err = doc.GenMarkdownTree(rootCmd, docDir)
				util.Check(err)
				util.PrintSuccess(fmt.Sprintf("markdown documentation generated at %s", docDir))
			case "man":
				manDir := path.Join(docDir, "man/")
				err = os.MkdirAll(manDir, util.DirPerm)
				util.Check(err)
				header := &doc.GenManHeader{
					Title:   util.GetCliBinaryName(),
					Section: "1",
				}
				err := doc.GenManTree(rootCmd, header, manDir)
				util.Check(err)
				util.PrintSuccess(fmt.Sprintf("man documentation generated at %s", manDir))
			default:
				util.Check(util.NewNotFoundError(fmt.Sprintf("%s documentation format not supported for documentation generation\n Supported types are MAN and MD", t)))
			}
		},
	}

	cmd.Flags().StringVarP(&docDir, "output-dir", "o", "", "Output dir path")
	return cmd
}
