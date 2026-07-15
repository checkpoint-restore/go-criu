package cli

import (
	"errors"
	"fmt"

	"github.com/checkpoint-restore/go-criu/v8/crit"
	"github.com/spf13/cobra"
)

var decompressInPlace bool

var decompressCmd = &cobra.Command{
	Use:   "decompress DIR",
	Short: "Decompress memory pages in a checkpoint directory",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := terminationSignalContext(cmd.Context())
		defer stop()
		result, err := crit.DecompressCheckpoint(ctx, args[0], crit.DecompressOptions{
			InPlace: decompressInPlace,
		})
		if err != nil {
			if errors.Is(err, crit.ErrCheckpointCompressionCleanup) {
				return fmt.Errorf("checkpoint was decompressed, but rollback cleanup failed: %w", err)
			}
			return fmt.Errorf("failed to decompress checkpoint: %w", err)
		}
		out := cmd.OutOrStdout()
		if result.AlreadyInRequestedState {
			_, err = fmt.Fprintf(out, "Checkpoint in %s is already decompressed\n", args[0])
			return err
		}
		if !result.Changed {
			_, err = fmt.Fprintf(out, "No pagemap files found in %s\n", args[0])
			return err
		}
		if _, err = fmt.Fprintf(out, "Decompressing checkpoint in %s\n", args[0]); err != nil {
			return err
		}
		for _, stats := range result.Pagemaps {
			if _, err = fmt.Fprintf(out, "  %s: %d pages (%dK -> %dK)\n",
				stats.Pagemap, stats.Pages, stats.InputBytes/1024, stats.OutputBytes/1024); err != nil {
				return err
			}
		}
		_, err = fmt.Fprintln(out, "Done")
		return err
	},
}

func init() {
	decompressCmd.Flags().BoolVar(&decompressInPlace, "in-place", false,
		"Skip creating backup files")
	rootCmd.AddCommand(decompressCmd)
}
