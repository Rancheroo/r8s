package cmd

import (
	"fmt"
	"strings"

	"github.com/Rancheroo/r8s/internal/bundle"
	"github.com/spf13/cobra"
)

var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Work with support bundles",
	Long: `Work with support bundles for offline analysis.

Support bundles are tar.gz archives containing cluster diagnostics,
logs, and configuration files. This command allows you to import
and analyze bundles without needing a live cluster connection.`,
}

var (
	bundlePath    string
	bundleMaxSize int64
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a support bundle",
	Long: `Import a support bundle from a tar.gz file.

The bundle will be extracted and analyzed. You can then use r8s
to browse the bundle contents offline, similar to browsing a live cluster.

Example:
  r8s bundle import --path=bundle.tar.gz
  r8s bundle import --path=bundle.tar.gz --limit=20
`,
	RunE: runImport,
}

func init() {
	rootCmd.AddCommand(bundleCmd)
	bundleCmd.AddCommand(importCmd)

	// Import subcommand flags
	importCmd.Flags().StringVarP(&bundlePath, "path", "p", "", "Path to bundle tar.gz file (required)")
	importCmd.Flags().Int64VarP(&bundleMaxSize, "limit", "l", 10, "Maximum bundle size in MB (default 10MB)")
	importCmd.MarkFlagRequired("path")
}

func runImport(cmd *cobra.Command, args []string) error {
	fmt.Printf("Importing bundle: %s\n", bundlePath)
	fmt.Printf("Size limit: %dMB\n\n", bundleMaxSize)

	// Create import options
	opts := bundle.ImportOptions{
		Path:    bundlePath,
		MaxSize: bundleMaxSize * 1024 * 1024, // Convert MB to bytes
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
	fmt.Println("Bundle Import Successful!")
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
	fmt.Println("\n✓ Bundle successfully imported and ready for analysis!")
	fmt.Println("\nNext steps:")
	fmt.Println("  • Use 'r8s' to browse the bundle in TUI mode")
	fmt.Println("  • Bundle will remain extracted until system cleanup")
	fmt.Printf("  • Extraction location: %s\n", b.ExtractPath)

	return nil
}
