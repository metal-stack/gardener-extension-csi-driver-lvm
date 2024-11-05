package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"
)

var log = runtimelog.Log.WithName("gardener-extension-csi-driver-lvm")

func NewControllerManagerCommand(ctx context.Context) *cobra.Command {
	options := NewOptions()
	cmd := &cobra.Command{
		Use:           "gardener-extension-csi-driver-lvm",
		Short:         "provides csi-driver-lvm for shoot cluster",
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			err := options.optionAggregator.Complete()
			if err != nil {
				return fmt.Errorf("error completing options: %w", err)
			}
			if err := options.heartbeatOptions.Validate(); err != nil {
				return err
			}

			cmd.SilenceUsage = true

			return options.run(ctx)
		},
	}

	options.optionAggregator.AddFlags(cmd.Flags())

	return cmd
}
