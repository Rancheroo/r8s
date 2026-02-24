// Package ai provides pattern matching and root cause analysis for Kubernetes issues.
// Sprint 11 Day 12: Parallel Analyzer - Performance optimization with goroutines
package ai

import (
	"context"
	"runtime"
	"sync"
	"time"
)

// ParallelAnalyzer provides concurrent pattern matching for improved performance
type ParallelAnalyzer struct {
	registry   *PatternRegistryV2
	generator  *HintGeneratorV2
	workerCount int
}

// NewParallelAnalyzer creates a parallel analyzer with optimal worker count
func NewParallelAnalyzer() *ParallelAnalyzer {
	// Use num CPUs or 4, whichever is higher
	workers := runtime.NumCPU()
	if workers < 4 {
		workers = 4
	}
	
	return &ParallelAnalyzer{
		registry:    NewRegistryV2(),
		generator:   NewHintGenerator(),
		workerCount: workers,
	}
}

// SetWorkerCount allows tuning the number of parallel workers
func (pa *ParallelAnalyzer) SetWorkerCount(n int) {
	if n > 0 {
		pa.workerCount = n
	}
}

// PatternMatchTask represents a unit of work for a worker
type PatternMatchTask struct {
	Pattern  PatternV2
	Content  string
	FileName string
}

// PatternMatchResult represents the outcome of a match task
type PatternMatchResult struct {
	Matches []MatchResultV2
	Pattern PatternV2
}

// AnalyzeParallel performs concurrent pattern matching on content
func (pa *ParallelAnalyzer) AnalyzeParallel(ctx context.Context, content map[string]string, opts AnalysisOptions) (*AnalysisResult, error) {
	startTime := time.Now()
	
	// Create worker pool
	tasks := make(chan PatternMatchTask, len(pa.registry.GetAll())*len(content))
	results := make(chan PatternMatchResult, len(pa.registry.GetAll())*len(content))
	
	var wg sync.WaitGroup
	
	// Start workers
	for i := 0; i < pa.workerCount; i++ {
		wg.Add(1)
		go pa.worker(ctx, &wg, tasks, results)
	}
	
	// Close results channel when workers done
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Submit tasks for each file + pattern combination
	totalTasks := 0
	for fileName, fileContent := range content {
		for _, pattern := range pa.registry.GetAll() {
			if pa.shouldAnalyzePattern(pattern, opts) {
				tasks <- PatternMatchTask{
					Pattern:  pattern,
					Content:  fileContent,
					FileName: fileName,
				}
				totalTasks++
			}
		}
	}
	close(tasks)
	
	// Collect results
	var allMatches []MatchResultV2
	for result := range results {
		allMatches = append(allMatches, result.Matches...)
	}
	
	// Generate hints
	hints := pa.generator.GenerateAll(allMatches, pa.registry)
	
	// Detect correlations
	correlations := pa.detectCorrelations(allMatches)
	
	// Build summary
	summary := pa.buildSummary(allMatches, correlations)
	
	endTime := time.Now()
	
	return &AnalysisResult{
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Patterns:     allMatches,
		Hints:        hints,
		Correlations: correlations,
		Summary:      summary,
	}, nil
}

// worker processes pattern matching tasks concurrently
func (pa *ParallelAnalyzer) worker(ctx context.Context, wg *sync.WaitGroup, tasks <-chan PatternMatchTask, results chan<- PatternMatchResult) {
	defer wg.Done()
	
	for task := range tasks {
		select {
		case <-ctx.Done():
			return
		default:
			matcher := NewMatcherV2(task.Pattern)
			matches := matcher.Match(task.Content)
			
			// Add file name to each match
			for i := range matches {
				if matches[i].Matched {
					if matches[i].Metadata == nil {
						matches[i].Metadata = make(map[string]string)
					}
					matches[i].Metadata["SourceFile"] = task.FileName
				}
			}
			
			results <- PatternMatchResult{
				Matches: matches,
				Pattern: task.Pattern,
			}
		}
	}
}

// shouldAnalyzePattern determines if a pattern should be analyzed based on options
func (pa *ParallelAnalyzer) shouldAnalyzePattern(p PatternV2, opts AnalysisOptions) bool {
	// Check severity filter
	if opts.MinSeverity != "" {
		severityOrder := map[Severity]int{
			SeverityInfo:     0,
			SeverityWarning:  1,
			SeverityCritical: 2,
		}
		if severityOrder[p.Severity] < severityOrder[opts.MinSeverity] {
			return false
		}
	}
	
	return true
}

// detectCorrelations finds all correlations between matched patterns
func (pa *ParallelAnalyzer) detectCorrelations(matches []MatchResultV2) []CorrelationMatch {
	var correlations []CorrelationMatch
	matchedIDs := make(map[string]bool)

	for _, match := range matches {
		if match.Matched {
			matchedIDs[match.PatternID] = true
		}
	}

	for _, match := range matches {
		if !match.Matched {
			continue
		}
		
		pattern, found := pa.registry.GetByID(match.PatternID)
		if !found {
			continue
		}

		for _, corr := range pattern.Correlations {
			if matchedIDs[corr.PatternID] {
				correlations = append(correlations, CorrelationMatch{
					PatternID1: match.PatternID,
					PatternID2: corr.PatternID,
					Message:    corr.Message,
				})
			}
		}
	}

	return correlations
}

// buildSummary creates analysis summary statistics
func (pa *ParallelAnalyzer) buildSummary(matches []MatchResultV2, correlations []CorrelationMatch) AnalysisSummary {
	summary := AnalysisSummary{
		TotalPatterns: len(pa.registry.GetAll()),
		MatchesFound:  0,
		Correlations:  len(correlations),
	}

	for _, match := range matches {
		if match.Matched {
			summary.MatchesFound++
			switch match.Severity {
			case SeverityCritical:
				summary.CriticalIssues++
			case SeverityWarning:
				summary.WarningIssues++
			case SeverityInfo:
				summary.InfoIssues++
			}
		}
	}

	return summary
}

// BenchmarkResult represents performance benchmark results
type BenchmarkResult struct {
	Duration       time.Duration
	PatternsPerSec float64
	FilesPerSec    float64
	WorkersUsed    int
}

// Benchmark runs a performance benchmark
func (pa *ParallelAnalyzer) Benchmark(content map[string]string, iterations int) BenchmarkResult {
	start := time.Now()
	
	for i := 0; i < iterations; i++ {
		ctx := context.Background()
		_, _ = pa.AnalyzeParallel(ctx, content, AnalysisOptions{})
	}
	
	duration := time.Since(start)
	
	return BenchmarkResult{
		Duration:       duration,
		PatternsPerSec: float64(len(pa.registry.GetAll())*iterations) / duration.Seconds(),
		FilesPerSec:    float64(len(content)*iterations) / duration.Seconds(),
		WorkersUsed:    pa.workerCount,
	}
}

// ProgressFunc is called during analysis to report progress
type ProgressFunc func(completed, total int, currentFile string)

// AnalyzeParallelWithProgress runs analysis with progress callback
func (pa *ParallelAnalyzer) AnalyzeParallelWithProgress(
	ctx context.Context,
	content map[string]string,
	opts AnalysisOptions,
	progress ProgressFunc,
) (*AnalysisResult, error) {
	
	if progress == nil {
		// No progress callback, use standard parallel analysis
		return pa.AnalyzeParallel(ctx, content, opts)
	}
	
	startTime := time.Now()
	
	// Calculate total tasks
	totalTasks := len(pa.registry.GetAll()) * len(content)
	completedTasks := 0
	var progressMu sync.Mutex
	
	// Create worker pool with progress tracking
	tasks := make(chan PatternMatchTask, totalTasks)
	results := make(chan PatternMatchResult, totalTasks)
	
	var wg sync.WaitGroup
	
	// Start workers
	for i := 0; i < pa.workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range tasks {
				select {
				case <-ctx.Done():
					return
				default:
					matcher := NewMatcherV2(task.Pattern)
					matches := matcher.Match(task.Content)
					
					// Add file name to each match
					for i := range matches {
						if matches[i].Matched {
							if matches[i].Metadata == nil {
								matches[i].Metadata = make(map[string]string)
							}
							matches[i].Metadata["SourceFile"] = task.FileName
						}
					}
					
					results <- PatternMatchResult{
						Matches: matches,
						Pattern: task.Pattern,
					}
					
					// Report progress
					progressMu.Lock()
					completedTasks++
					progress(completedTasks, totalTasks, task.FileName)
					progressMu.Unlock()
				}
			}
		}()
	}
	
	// Close results when done
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Submit tasks
	for fileName, fileContent := range content {
		for _, pattern := range pa.registry.GetAll() {
			if pa.shouldAnalyzePattern(pattern, opts) {
				tasks <- PatternMatchTask{
					Pattern:  pattern,
					Content:  fileContent,
					FileName: fileName,
				}
			}
		}
	}
	close(tasks)
	
	// Collect results
	var allMatches []MatchResultV2
	for result := range results {
		allMatches = append(allMatches, result.Matches...)
	}
	
	// Generate hints
	hints := pa.generator.GenerateAll(allMatches, pa.registry)
	
	// Detect correlations
	correlations := pa.detectCorrelations(allMatches)
	
	// Build summary
	summary := pa.buildSummary(allMatches, correlations)
	
	endTime := time.Now()
	
	return &AnalysisResult{
		StartTime:    startTime,
		EndTime:      endTime,
		Duration:     endTime.Sub(startTime),
		Patterns:     allMatches,
		Hints:        hints,
		Correlations: correlations,
		Summary:      summary,
	}, nil
}