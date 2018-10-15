package operator

import (
	"github.com/spf13/cobra"

	"github.com/openshift/library-go/pkg/controller/controllercmd"

	"github.com/enj/example-operator/pkg/starter"
	"github.com/enj/example-operator/pkg/version"
)

func NewOperator() *cobra.Command {
	cmd := controllercmd.
		NewControllerCommandConfig("example-operator", version.Get(), starter.RunOperator).
		NewCommand()
	cmd.Use = "operator"
	cmd.Short = "Start the Example Operator"

	return cmd
}
