package cmd

import (
	"fmt"
	"strings"

	"github.com/Rancheroo/r8s/internal/bundle"
	"github.com/spf13/cobra"
)

var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "View support bundle information",
	Long: `View information about support bundles without launching TUI.

Support bundles are RKE2 diagnostic archives containing cluster logs,
configurations, and kubectl resource dumps.

RECOMMENDED WORKFLOW:
  # Just point at the bundle and it works!
  r8s ./extracted-bundle-folder/
  r8s ./support-bundle.tar.gz

  # Or use the TUI command explicitly
  r8s tui --bundle=./extracted-folder/

COMMANDS:
  info   View bundle metadata and statistics (for inspection/validation)

TIP: For interactive analysis, just run 'r8s ./bundle-path' directly!`,
}

var (
	bundleMaxSize int64
)

var infoCmd = &cobra.Command{
	Use:   "info [path]",
	Short: "Display bundle information",
	Long: `Display metadata and statistics about a support bundle.

This command shows bundle info without launching the TUI.
Useful for quick inspection or CI/CD validation.

For interactive analysis, just run: r8s ./bundle-path

EXAMPLES:
  # View bundle info
  r8s bundle info ./w-guard-wg-cp-xyz/
  r8s bundle info bundle.tar.gz

  # Set custom size limit
  r8s bundle info bundle.tar.gz --limit=100`,
	Args: cobra.ExactArgs(1),
	RunE: runBundleInfo,
}

func init() {
	rootCmd.AddCommand(bundleCmd)
	bundleCmd.AddCommand(infoCmd)

	// Info command flags
	infoCmd.Flags().Int64VarP(&bundleMaxSize, "limit", "l", 50, "Maximum bundle size in MB (default 50, use 0 for unlimited)")
}

func runBundleInfo(cmd *cobra.Command, args []string) error {
	// Get bundle path from positional argument
	bundlePath := args[0]

	fmt.Printf("Analyzing bundle: %s\n", bundlePath)

	// Display size limit (show default if 0 or negative)
	if bundleMaxSize <= 0 {
		fmt.Printf("Size limit: 50MB (default)\n\n")
	} else if bundleMaxSize == 50 {
		fmt.Printf("Size limit: 50MB (default)\n\n")
	} else {
		fmt.Printf("Size limit: %dMB\n\n", bundleMaxSize)
	}

	// Create import options
	opts := bundle.ImportOptions{
		Path:    bundlePath,
		MaxSize: bundleMaxSize * 1024 * 1024, // Convert MB to bytes
		Verbose: verbose,                     // Pass verbose flag from root
	}

	// Load the bundle
	fmt.Println("Extracting bundle...")
	b, err := bundle.Load(opts)
	if err != nil {
		return fmt.Errorf("failed to load bundle: %w", err)
	}
	defer b.Close()

	// Display bundle information
	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Println("Bundle Analysis Complete!")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("\nNode Name:     %s\n", b.Manifest.NodeName)
	fmt.Printf("Bundle Type:   %s\n", b.Manifest.BundleType)
	fmt.Printf("RKE2 Version:  %s\n", b.Manifest.RKE2Version)
	fmt.Printf("K8s Version:   %s\n", b.Manifest.K8sVersion)
	fmt.Printf("Collected:     %s\n", b.Manifest.CollectedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("\nBundle Size:   %.2f MB\n", float64(b.Size)/(1024*1024))
	fmt.Printf("Files:         %d\n", b.Manifest.FileCount)
	fmt.Printf("Pods Found:    %d\n", len(b.Pods))
	fmt.Printf("Log Files:     %d\n", len(b.LogFiles))

	// Show pod summary
	if len(b.Pods) > 0 {
		fmt.Println("\nPod Summary by Namespace:")
		fmt.Println(strings.Repeat("─", 60))

		// Group pods by namespace
		nsPods := make(map[string]int)
		for _, pod := range b.Pods {
			nsPods[pod.Namespace]++
		}

		// Display counts
		for ns, count := range nsPods {
			fmt.Printf("  %-30s %d pods\n", ns, count)
		}
	}

	// Show log file types
	if len(b.LogFiles) > 0 {
		fmt.Println("\nLog Files by Type:")
		fmt.Println(strings.Repeat("─", 60))

		// Group logs by type
		logTypes := make(map[bundle.LogType]int)
		for _, log := range b.LogFiles {
			logTypes[log.Type]++
		}

		// Display counts
		for logType, count := range logTypes {
			fmt.Printf("  %-30s %d files\n", logType, count)
		}
	}

	fmt.Println("\n" + strings.Repeat("─", 60))
	fmt.Println("\n✓ Bundle loaded successfully and ready for analysis!")
	fmt.Println("\nTo analyze interactively:")
	fmt.Printf("  r8s %s\n", bundlePath)

	return nil
}
