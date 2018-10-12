package operator

import (
	"github.com/spf13/cobra"

	"github.com/enj/example-operator/pkg/operator"
	"github.com/enj/example-operator/pkg/version"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
)

func NewOperator() *cobra.Command {
	cmd := controllercmd.
		NewControllerCommandConfig("example-operator", version.Get(), operator.RunOperator).
		NewCommand()
	cmd.Use = "operator"
	cmd.Short = "Start the Example Operator"

	return cmd
}
