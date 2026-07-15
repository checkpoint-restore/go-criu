package cli

import (
	"errors"
	"fmt"

	"github.com/checkpoint-restore/go-criu/v8/crit"
	"github.com/spf13/cobra"
)

var (
	compressInPlace      bool
	compressAcceleration int
)

var compressCmd = &cobra.Command{
	Use:   "compress DIR",
	Short: "Compress memory pages in a checkpoint directory",
	Long: `Compress memory pages in a checkpoint directory.

For an incremental checkpoint chain, DIR must be its newest layer: no other
checkpoint may use DIR as its parent. A standalone checkpoint also qualifies.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if compressAcceleration != 1 {
			return fmt.Errorf("failed to compress checkpoint: %w: %d (only 1 is supported)",
				crit.ErrUnsupportedAcceleration, compressAcceleration)
		}
		ctx, stop := terminationSignalContext(cmd.Context())
		defer stop()
		result, err := crit.CompressCheckpoint(ctx, args[0], crit.CompressOptions{
			InPlace:      compressInPlace,
			Acceleration: compressAcceleration,
		})
		if err != nil {
			if errors.Is(err, crit.ErrCheckpointCompressionCleanup) {
				return fmt.Errorf("checkpoint was compressed, but rollback cleanup failed: %w", err)
			}
			return fmt.Errorf("failed to compress checkpoint: %w", err)
		}
		out := cmd.OutOrStdout()
		if result.AlreadyInRequestedState {
			_, err = fmt.Fprintf(out, "Checkpoint in %s is already compressed\n", args[0])
			return err
		}
		if !result.Changed {
			_, err = fmt.Fprintf(out, "No pagemap files found in %s\n", args[0])
			return err
		}
		if _, err = fmt.Fprintf(out, "Compressing checkpoint in %s\n", args[0]); err != nil {
			return err
		}
		for _, stats := range result.Pagemaps {
			saved := float64(0)
			if stats.InputBytes != 0 {
				saved = (1 - float64(stats.OutputBytes)/float64(stats.InputBytes)) * 100
			}
			if _, err = fmt.Fprintf(out, "  %s: %d pages (%dK -> %dK, %.1f%% saved)\n",
				stats.Pagemap, stats.Pages, stats.InputBytes/1024, stats.OutputBytes/1024, saved); err != nil {
				return err
			}
		}
		_, err = fmt.Fprintln(out, "Done")
		return err
	},
}

func init() {
	compressCmd.Flags().BoolVar(&compressInPlace, "in-place", false,
		"Skip creating backup files")
	compressCmd.Flags().IntVar(&compressAcceleration, "acceleration", 1,
		"LZ4 acceleration (only 1 is currently supported)")
	rootCmd.AddCommand(compressCmd)
}
