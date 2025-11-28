package bundle

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Extract extracts a tar.gz bundle to a destination directory.
// It enforces size limits and validates the extraction process.
func Extract(bundlePath string, opts ImportOptions) (string, error) {
	// Open the bundle file
	file, err := os.Open(bundlePath)
	if err != nil {
		return "", fmt.Errorf("failed to open bundle: %w", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to stat bundle: %w", err)
	}

	// Check compressed size limit (rough estimate)
	if opts.MaxSize > 0 && stat.Size() > opts.MaxSize {
		return "", fmt.Errorf("bundle file size (%d bytes) exceeds limit (%d bytes)",
			stat.Size(), opts.MaxSize)
	}

	// Create extraction directory
	extractPath := opts.ExtractTo
	if extractPath == "" {
		extractPath, err = os.MkdirTemp("", "r8s-bundle-*")
		if err != nil {
			return "", fmt.Errorf("failed to create temp directory: %w", err)
		}
	}

	// Create gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		os.RemoveAll(extractPath)
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	// Create tar reader
	tr := tar.NewReader(gzr)

	// Track total extracted size
	var totalExtracted int64
	fileCount := 0

	// Extract all files
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			os.RemoveAll(extractPath)
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}

		// Validate header name (prevent directory traversal)
		if strings.Contains(header.Name, "..") {
			os.RemoveAll(extractPath)
			return "", fmt.Errorf("invalid file path in bundle: %s", header.Name)
		}

		// Build target path
		target := filepath.Join(extractPath, header.Name)

		// Check uncompressed size limit
		totalExtracted += header.Size
		if opts.MaxSize > 0 && totalExtracted > opts.MaxSize {
			os.RemoveAll(extractPath)
			// Calculate sizes in MB for user-friendly message
			sizeMB := float64(totalExtracted) / (1024 * 1024)
			limitMB := float64(opts.MaxSize) / (1024 * 1024)

			if opts.Verbose {
				return "", fmt.Errorf("bundle uncompressed size (%.1f MB) exceeds limit (%.1f MB)\n\n"+
					"The bundle is too large for the current size limit.\n\n"+
					"SOLUTION:\n"+
					"  Increase the limit with --limit flag:\n"+
					"  r8s bundle import --path=%s --limit=%d\n\n"+
					"DETAILS:\n"+
					"  Current limit: %.1f MB\n"+
					"  Bundle size:   %.1f MB\n"+
					"  Suggested:     --limit=%d (or higher)\n\n"+
					"SAFETY NOTES:\n"+
					"  • Size limits prevent system OOM (out of memory)\n"+
					"  • Reasonable limit: 100-500 MB for typical bundles\n"+
					"  • Maximum safe limit depends on available RAM\n"+
					"  • Use --limit=0 to disable (not recommended for large files)",
					sizeMB, limitMB, filepath.Base(bundlePath), int(sizeMB)+10,
					limitMB, sizeMB, int(sizeMB)+10)
			}

			return "", fmt.Errorf("bundle size (%.1f MB) exceeds limit (%.1f MB)\n"+
				"Solution: Use --limit=%d to increase (e.g. 'r8s bundle import --path=bundle.tar.gz --limit=%d')",
				sizeMB, limitMB, int(sizeMB)+10, int(sizeMB)+10)
		}

		// Handle different file types
		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(target, 0755); err != nil {
				os.RemoveAll(extractPath)
				return "", fmt.Errorf("failed to create directory %s: %w", target, err)
			}

		case tar.TypeReg:
			// Extract regular file
			if err := extractFile(tr, target, header); err != nil {
				os.RemoveAll(extractPath)
				return "", fmt.Errorf("failed to extract file %s: %w", target, err)
			}
			fileCount++

		case tar.TypeSymlink:
			// FIX BUG #1: Validate symlink target to prevent panic on malicious/broken links
			// Skip symlinks entirely - they can cause traversal issues and aren't needed for bundle data
			if opts.Verbose {
				fmt.Printf("⚠ Skipping symlink: %s -> %s (symlinks not supported in bundles)\n",
					header.Name, header.Linkname)
			}
			continue

		default:
			// Skip other types (block devices, etc.)
			continue
		}
	}

	return extractPath, nil
}

// extractFile extracts a single file from the tar reader.
func extractFile(tr *tar.Reader, target string, header *tar.Header) error {
	// Ensure parent directory exists
	dir := filepath.Dir(target)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create the file
	file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy contents
	if _, err := io.Copy(file, tr); err != nil {
		return err
	}

	return nil
}

// Cleanup removes the extracted bundle directory.
func Cleanup(extractPath string) error {
	if extractPath == "" {
		return nil
	}
	return os.RemoveAll(extractPath)
}
