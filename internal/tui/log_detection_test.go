package tui

import "testing"

// TestLogLevelDetection tests the fixed WARN/ERROR detection logic
func TestLogLevelDetection(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantErr  bool
		wantWarn bool
	}{
		// Explicit WARN indicators should NOT be detected as errors
		{
			name:     "K8s WARN with failed keyword",
			line:     "W1204 09:15:57.456789 [WARN] Skipping failed migration, continuing",
			wantErr:  false,
			wantWarn: true,
		},
		{
			name:     "K8s WARN with denied keyword",
			line:     "W1204 09:15:58.345678 [WARN] Database unavailable, retrying in 5s",
			wantErr:  false,
			wantWarn: true,
		},
		{
			name:     "Bracket WARN with failed keyword",
			line:     "[WARN] Failed to list Services: connection refused",
			wantErr:  false,
			wantWarn: true,
		},

		// Explicit ERROR indicators
		{
			name:     "K8s ERROR format",
			line:     "E1204 09:15:58.234567 [ERROR] Cannot connect to database: connection refused at postgres:5432",
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "Bracket ERROR format",
			line:     "[ERROR] Failed to load initial config from /etc/app/config.yaml: file not found",
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "ERROR with FAILED keyword",
			line:     "E1204 09:16:03.456789 [ERROR] Database connection failed again: timeout after 5s",
			wantErr:  true,
			wantWarn: false,
		},

		// Pure keyword matching (no explicit indicator)
		{
			name:     "FAILED keyword without explicit indicator",
			line:     "Migration process failed at step 3",
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "PANIC keyword",
			line:     "panic: runtime error: index out of range",
			wantErr:  true,
			wantWarn: false,
		},
		{
			name:     "DEPRECATED keyword",
			line:     "API endpoint /v1/old is DEPRECATED, use /v2/new instead",
			wantErr:  false,
			wantWarn: true,
		},

		// INFO logs should not be detected as ERROR or WARN
		{
			name:     "INFO with failure word in message",
			line:     "I1127 00:44:40.586619 [INFO] Failed to read data from checkpoint - checkpoint not found",
			wantErr:  false,
			wantWarn: false,
		},
		{
			name:     "Plain INFO",
			line:     "I1127 00:44:40.476206 [INFO] Kubelet starting up...",
			wantErr:  false,
			wantWarn: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := isErrorLog(tt.line)
			gotWarn := isWarnLog(tt.line)

			if gotErr != tt.wantErr {
				t.Errorf("isErrorLog() = %v, want %v for line: %s", gotErr, tt.wantErr, tt.line)
			}
			if gotWarn != tt.wantWarn {
				t.Errorf("isWarnLog() = %v, want %v for line: %s", gotWarn, tt.wantWarn, tt.line)
			}

			// Ensure a line isn't detected as both ERROR and WARN
			if gotErr && gotWarn {
				t.Errorf("Line detected as both ERROR and WARN: %s", tt.line)
			}
		})
	}
}
