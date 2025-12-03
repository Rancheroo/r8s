package datasource

// NewEmbeddedDataSource creates a data source from the demo bundle
// This loads from the example bundle in the repo for demo mode
// TODO: In production, use go:embed to embed the bundle in the binary
func NewEmbeddedDataSource(verbose bool) (DataSource, error) {
	// Use the example bundle from the repo
	bundlePath := "example-log-bundle/w-guard-wg-cp-svtk6-lqtxw-2025-11-27_04_19_09"
	return NewBundleDataSource(bundlePath, verbose)
}
