package bigdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SparkBatchProcessor processes large conversation datasets using Apache Spark
type SparkBatchProcessor struct {
	sparkMasterURL string
	dataLakeClient *DataLakeClient
	outputPath     string
	logger         *logrus.Logger
	appName        string
	httpClient     *http.Client
}

// BatchJobType defines the type of batch processing job
type BatchJobType string

const (
	// BatchJobEntityExtraction extracts entities from historical conversations
	BatchJobEntityExtraction BatchJobType = "entity_extraction"

	// BatchJobRelationshipMining analyzes co-occurrence patterns
	BatchJobRelationshipMining BatchJobType = "relationship_mining"

	// BatchJobTopicModeling performs topic modeling (LDA/NMF)
	BatchJobTopicModeling BatchJobType = "topic_modeling"

	// BatchJobProviderPerformance aggregates provider scores
	BatchJobProviderPerformance BatchJobType = "provider_performance"

	// BatchJobDebateAnalysis analyzes multi-round debate patterns
	BatchJobDebateAnalysis BatchJobType = "debate_analysis"
)

// BatchParams defines parameters for batch processing job
type BatchParams struct {
	JobType    BatchJobType
	InputPath  string                 // S3/MinIO path to input data
	OutputPath string                 // S3/MinIO path for output
	StartDate  time.Time              // Optional date range filter
	EndDate    time.Time              // Optional date range filter
	Options    map[string]interface{} // Job-specific options
}

// BatchResult contains results from batch processing job
type BatchResult struct {
	JobID              string
	JobType            BatchJobType
	Status             string
	ProcessedRows      int64
	EntitiesExtracted  int64
	RelationshipsFound int64
	TopicsIdentified   int
	StartedAt          time.Time
	CompletedAt        time.Time
	DurationMs         int64
	OutputPath         string
	Metrics            map[string]interface{}
	ErrorMessage       string
}

// SparkJobConfig defines Spark job configuration
type SparkJobConfig struct {
	ExecutorMemory  string // e.g., "2g"
	ExecutorCores   int    // e.g., 2
	NumExecutors    int    // e.g., 4
	DriverMemory    string // e.g., "1g"
	DeployMode      string // "client" or "cluster"
	PythonFile      string // Path to PySpark script
	AdditionalArgs  []string
	EnvironmentVars map[string]string
}

// NewSparkBatchProcessor creates a new Spark batch processor
func NewSparkBatchProcessor(
	sparkMasterURL string,
	dataLakeClient *DataLakeClient,
	outputPath string,
	logger *logrus.Logger,
) *SparkBatchProcessor {
	if logger == nil {
		logger = logrus.New()
	}

	return &SparkBatchProcessor{
		sparkMasterURL: sparkMasterURL,
		dataLakeClient: dataLakeClient,
		outputPath:     outputPath,
		logger:         logger,
		appName:        "HelixAgent-BigData",
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

// ProcessConversationDataset executes a batch processing job
func (sbp *SparkBatchProcessor) ProcessConversationDataset(
	ctx context.Context,
	params BatchParams,
) (*BatchResult, error) {
	startTime := time.Now()

	sbp.logger.WithFields(logrus.Fields{
		"job_type":    params.JobType,
		"input_path":  params.InputPath,
		"output_path": params.OutputPath,
	}).Info("Starting Spark batch processing job")

	// Create job-specific Spark configuration
	config := sbp.createJobConfig(params)

	// Validate input path exists
	exists, err := sbp.dataLakeClient.PathExists(ctx, params.InputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to check input path: %w", err)
	}
	if !exists {
		return nil, fmt.Errorf("input path does not exist: %s", params.InputPath)
	}

	// Submit Spark job
	result, err := sbp.submitSparkJob(ctx, config, params)
	if err != nil {
		return &BatchResult{
			JobType:      params.JobType,
			Status:       "failed",
			StartedAt:    startTime,
			CompletedAt:  time.Now(),
			DurationMs:   time.Since(startTime).Milliseconds(),
			ErrorMessage: err.Error(),
		}, err
	}

	result.StartedAt = startTime
	result.CompletedAt = time.Now()
	result.DurationMs = time.Since(startTime).Milliseconds()

	sbp.logger.WithFields(logrus.Fields{
		"job_id":        result.JobID,
		"status":        result.Status,
		"duration_ms":   result.DurationMs,
		"rows":          result.ProcessedRows,
		"entities":      result.EntitiesExtracted,
		"relationships": result.RelationshipsFound,
	}).Info("Spark batch processing job completed")

	return result, nil
}

// createJobConfig creates Spark configuration for specific job type
func (sbp *SparkBatchProcessor) createJobConfig(params BatchParams) *SparkJobConfig {
	config := &SparkJobConfig{
		ExecutorMemory: "2g",
		ExecutorCores:  2,
		NumExecutors:   4,
		DriverMemory:   "1g",
		DeployMode:     "client",
		EnvironmentVars: map[string]string{
			"SPARK_HOME": "/opt/spark",
		},
	}

	// Job-specific configuration
	switch params.JobType {
	case BatchJobEntityExtraction:
		config.PythonFile = filepath.Join(sbp.outputPath, "scripts", "entity_extraction.py")
		config.ExecutorMemory = "4g"
		config.NumExecutors = 8

	case BatchJobRelationshipMining:
		config.PythonFile = filepath.Join(sbp.outputPath, "scripts", "relationship_mining.py")
		config.ExecutorMemory = "4g"
		config.NumExecutors = 8

	case BatchJobTopicModeling:
		config.PythonFile = filepath.Join(sbp.outputPath, "scripts", "topic_modeling.py")
		config.ExecutorMemory = "8g"
		config.NumExecutors = 4

	case BatchJobProviderPerformance:
		config.PythonFile = filepath.Join(sbp.outputPath, "scripts", "provider_performance.py")
		config.ExecutorMemory = "2g"
		config.NumExecutors = 4

	case BatchJobDebateAnalysis:
		config.PythonFile = filepath.Join(sbp.outputPath, "scripts", "debate_analysis.py")
		config.ExecutorMemory = "4g"
		config.NumExecutors = 6
	}

	return config
}

// submitSparkJob submits a Spark job and monitors execution
func (sbp *SparkBatchProcessor) submitSparkJob(
	ctx context.Context,
	config *SparkJobConfig,
	params BatchParams,
) (*BatchResult, error) {
	// Build spark-submit command
	args := []string{
		"--master", sbp.sparkMasterURL,
		"--deploy-mode", config.DeployMode,
		"--executor-memory", config.ExecutorMemory,
		"--executor-cores", fmt.Sprintf("%d", config.ExecutorCores),
		"--num-executors", fmt.Sprintf("%d", config.NumExecutors),
		"--driver-memory", config.DriverMemory,
		"--name", fmt.Sprintf("%s-%s", sbp.appName, params.JobType),
	}

	// Add Python packages if needed
	args = append(args, "--packages", "org.apache.spark:spark-sql-kafka-0-10_2.12:3.5.0")

	// Add Python file
	args = append(args, config.PythonFile)

	// Add job arguments
	jobArgs, err := sbp.buildJobArgs(params)
	if err != nil {
		return nil, fmt.Errorf("failed to build job arguments: %w", err)
	}
	args = append(args, jobArgs...)

	sbp.logger.WithFields(logrus.Fields{
		"command": "spark-submit",
		"args":    args,
	}).Debug("Submitting Spark job")

	// Execute spark-submit
	cmd := exec.CommandContext(ctx, "spark-submit", args...)

	// Set environment variables
	for key, value := range config.EnvironmentVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		sbp.logger.WithError(err).WithField("output", string(output)).Error("Spark job failed")
		return nil, fmt.Errorf("spark job failed: %w\nOutput: %s", err, string(output))
	}

	// Parse job output to extract results
	result, err := sbp.parseJobOutput(string(output), params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse job output: %w", err)
	}

	return result, nil
}

// buildJobArgs builds command-line arguments for Spark job
func (sbp *SparkBatchProcessor) buildJobArgs(params BatchParams) ([]string, error) {
	args := []string{
		"--input-path", params.InputPath,
		"--output-path", params.OutputPath,
		"--job-type", string(params.JobType),
	}

	// Add date range if specified
	if !params.StartDate.IsZero() {
		args = append(args, "--start-date", params.StartDate.Format("2006-01-02"))
	}
	if !params.EndDate.IsZero() {
		args = append(args, "--end-date", params.EndDate.Format("2006-01-02"))
	}

	// Add job-specific options
	if len(params.Options) > 0 {
		optionsJSON, err := json.Marshal(params.Options)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal options: %w", err)
		}
		args = append(args, "--options", string(optionsJSON))
	}

	return args, nil
}

// parseJobOutput parses Spark job output to extract results
func (sbp *SparkBatchProcessor) parseJobOutput(output string, params BatchParams) (*BatchResult, error) {
	result := &BatchResult{
		JobID:   fmt.Sprintf("job-%d", time.Now().UnixNano()),
		JobType: params.JobType,
		Status:  "completed",
		Metrics: make(map[string]interface{}),
	}

	// Try to parse JSON output from Spark
	if parsed := sbp.parseJSONOutput(output, params); parsed != nil {
		result = parsed
	} else {
		// If we cannot parse JSON output, return an error
		return nil, fmt.Errorf("failed to parse Spark job output: no valid JSON found")
	}

	result.OutputPath = params.OutputPath

	sbp.logger.WithFields(logrus.Fields{
		"job_id": result.JobID,
		"status": result.Status,
	}).Debug("Parsed Spark job output")

	return result, nil
}

// parseJSONOutput attempts to parse JSON lines from Spark job output
func (sbp *SparkBatchProcessor) parseJSONOutput(output string, params BatchParams) *BatchResult {
	// Split output by lines and look for JSON objects
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || line[0] != '{' {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			continue
		}

		// Check if this looks like a Spark job result
		if _, hasStatus := data["status"]; !hasStatus {
			continue
		}

		result := &BatchResult{
			JobID:   fmt.Sprintf("job-%d", time.Now().UnixNano()),
			JobType: params.JobType,
			Status:  "completed",
			Metrics: make(map[string]interface{}),
		}

		// Extract common fields
		if val, ok := data["processed_rows"]; ok {
			if f, ok := val.(float64); ok {
				result.ProcessedRows = int64(f)
			}
		}
		if val, ok := data["entities_extracted"]; ok {
			if f, ok := val.(float64); ok {
				result.EntitiesExtracted = int64(f)
			}
		}
		if val, ok := data["relationships_found"]; ok {
			if f, ok := val.(float64); ok {
				result.RelationshipsFound = int64(f)
			}
		}
		if val, ok := data["topics_identified"]; ok {
			if f, ok := val.(float64); ok {
				result.TopicsIdentified = int(f)
			}
		}

		// Copy all other fields to metrics
		for key, val := range data {
			switch key {
			case "processed_rows", "entities_extracted", "relationships_found", "topics_identified", "status":
				// Already handled
			default:
				result.Metrics[key] = val
			}
		}

		result.OutputPath = params.OutputPath
		sbp.logger.WithField("job_id", result.JobID).Debug("Parsed JSON output from Spark job")
		return result
	}

	// No valid JSON found
	return nil
}

// getSparkRESTBaseURL returns the base URL for Spark REST API
func (sbp *SparkBatchProcessor) getSparkRESTBaseURL() (string, error) {
	// Parse sparkMasterURL to extract host
	// Examples: spark://localhost:7077, local, yarn
	if sbp.sparkMasterURL == "" {
		return "", fmt.Errorf("spark master URL not configured")
	}

	// For simplicity, assume REST API runs on port 4040 on localhost
	// In production, this would parse the master URL and determine correct REST API endpoint
	return "http://localhost:4040", nil
}

// GetJobStatus retrieves status of a running Spark job
func (sbp *SparkBatchProcessor) GetJobStatus(ctx context.Context, jobID string) (*BatchResult, error) {
	// Get Spark REST API base URL
	baseURL, err := sbp.getSparkRESTBaseURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get Spark REST API URL: %w", err)
	}

	// Construct API endpoint
	// jobID should be Spark application ID
	apiURL := fmt.Sprintf("%s/api/v1/applications/%s/jobs", baseURL, url.PathEscape(jobID))

	sbp.logger.WithFields(logrus.Fields{
		"job_id":  jobID,
		"api_url": apiURL,
	}).Debug("Querying Spark job status via REST API")

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Execute request
	resp, err := sbp.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Spark REST API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Spark REST API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var sparkJobs []struct {
		JobID      int    `json:"jobId"`
		Status     string `json:"status"`
		StartTime  int64  `json:"startTime"`
		EndTime    int64  `json:"endTime"`
		NumTasks   int    `json:"numTasks"`
		NumActive  int    `json:"numActive"`
		NumFailed  int    `json:"numFailed"`
		NumKilled  int    `json:"numKilled"`
		NumSkipped int    `json:"numSkipped"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &sparkJobs); err != nil {
		return nil, fmt.Errorf("failed to parse Spark jobs response: %w", err)
	}

	// Convert to BatchResult
	if len(sparkJobs) == 0 {
		return nil, fmt.Errorf("no jobs found for application %s", jobID)
	}

	// Use first job for status (simplification - real implementation would aggregate)
	sparkJob := sparkJobs[0]
	result := &BatchResult{
		JobID:   jobID,
		JobType: BatchJobEntityExtraction, // Default - would need mapping
		Status:  strings.ToLower(sparkJob.Status),
		Metrics: map[string]interface{}{
			"spark_job_id":  sparkJob.JobID,
			"num_tasks":     sparkJob.NumTasks,
			"num_active":    sparkJob.NumActive,
			"num_failed":    sparkJob.NumFailed,
			"num_killed":    sparkJob.NumKilled,
			"num_skipped":   sparkJob.NumSkipped,
			"rest_api_used": true,
		},
	}

	// Set timestamps if available
	if sparkJob.StartTime > 0 {
		result.StartedAt = time.UnixMilli(sparkJob.StartTime)
	}
	if sparkJob.EndTime > 0 {
		result.CompletedAt = time.UnixMilli(sparkJob.EndTime)
		result.DurationMs = sparkJob.EndTime - sparkJob.StartTime
	}

	sbp.logger.WithFields(logrus.Fields{
		"job_id": jobID,
		"status": result.Status,
	}).Debug("Retrieved Spark job status from REST API")

	return result, nil
}

// CancelJob cancels a running Spark job
func (sbp *SparkBatchProcessor) CancelJob(ctx context.Context, jobID string) error {
	// Get Spark REST API base URL
	baseURL, err := sbp.getSparkRESTBaseURL()
	if err != nil {
		return fmt.Errorf("failed to get Spark REST API URL: %w", err)
	}

	// Construct API endpoint for application kill
	// jobID should be Spark application ID
	apiURL := fmt.Sprintf("%s/api/v1/applications/%s", baseURL, url.PathEscape(jobID))

	sbp.logger.WithFields(logrus.Fields{
		"job_id":  jobID,
		"api_url": apiURL,
	}).Debug("Cancelling Spark job via REST API")

	// Create DELETE request
	req, err := http.NewRequestWithContext(ctx, "DELETE", apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Execute request
	resp, err := sbp.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send cancellation request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	// Spark REST API returns 200 OK for successful kill
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Spark REST API returned status %d: %s", resp.StatusCode, string(body))
	}

	sbp.logger.WithField("job_id", jobID).Info("Spark job cancellation requested successfully")
	return nil
}

// getSparkHistoryBaseURL returns the base URL for Spark History Server REST API
func (sbp *SparkBatchProcessor) getSparkHistoryBaseURL() (string, error) {
	// Spark History Server typically runs on port 18080
	// In production, this would be configurable
	return "http://localhost:18080", nil
}

// ListCompletedJobs lists recently completed batch jobs
func (sbp *SparkBatchProcessor) ListCompletedJobs(
	ctx context.Context,
	limit int,
) ([]*BatchResult, error) {
	// Get Spark History Server base URL
	baseURL, err := sbp.getSparkHistoryBaseURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get Spark History Server URL: %w", err)
	}

	// Construct API endpoint for completed applications
	apiURL := fmt.Sprintf("%s/api/v1/applications", baseURL)

	sbp.logger.WithFields(logrus.Fields{
		"limit":   limit,
		"api_url": apiURL,
	}).Debug("Listing completed Spark jobs via History Server REST API")

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add query parameters for filtering
	q := req.URL.Query()
	q.Add("status", "completed")
	if limit > 0 {
		q.Add("limit", fmt.Sprintf("%d", limit))
	}
	req.URL.RawQuery = q.Encode()

	// Execute request
	resp, err := sbp.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Spark History Server: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Spark History Server returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var sparkApps []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Start   int64  `json:"startTime"`
		End     int64  `json:"endTime"`
		Status  string `json:"status"`
		User    string `json:"user"`
		Attempt int    `json:"attemptId"`
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, &sparkApps); err != nil {
		return nil, fmt.Errorf("failed to parse Spark applications response: %w", err)
	}

	// Convert to BatchResult
	jobs := make([]*BatchResult, 0, len(sparkApps))
	for _, app := range sparkApps {
		if strings.ToLower(app.Status) != "completed" {
			continue
		}

		// Determine job type from application name
		jobType := BatchJobEntityExtraction
		if strings.Contains(strings.ToLower(app.Name), "relationship") {
			jobType = BatchJobRelationshipMining
		} else if strings.Contains(strings.ToLower(app.Name), "topic") {
			jobType = BatchJobTopicModeling
		} else if strings.Contains(strings.ToLower(app.Name), "provider") {
			jobType = BatchJobProviderPerformance
		} else if strings.Contains(strings.ToLower(app.Name), "debate") {
			jobType = BatchJobDebateAnalysis
		}

		startedAt := time.UnixMilli(app.Start)
		completedAt := time.UnixMilli(app.End)
		durationMs := app.End - app.Start

		jobs = append(jobs, &BatchResult{
			JobID:         app.ID,
			JobType:       jobType,
			Status:        "completed",
			ProcessedRows: 0, // Would need additional API call to get actual metrics
			StartedAt:     startedAt,
			CompletedAt:   completedAt,
			DurationMs:    durationMs,
			OutputPath:    fmt.Sprintf("%s/results/%s", sbp.outputPath, app.ID),
			Metrics: map[string]interface{}{
				"application_name": app.Name,
				"user":             app.User,
				"attempt":          app.Attempt,
				"rest_api_used":    true,
			},
		})
	}

	sbp.logger.WithField("job_count", len(jobs)).Debug("Retrieved completed Spark jobs from History Server")
	return jobs, nil
}

// CleanupOldResults removes old batch processing results from data lake
func (sbp *SparkBatchProcessor) CleanupOldResults(
	ctx context.Context,
	olderThan time.Duration,
) (int, error) {
	if sbp.dataLakeClient == nil {
		return 0, fmt.Errorf("data lake client not configured")
	}

	thresholdTime := time.Now().Add(-olderThan)

	sbp.logger.WithFields(logrus.Fields{
		"older_than_hours": olderThan.Hours(),
		"output_path":      sbp.outputPath,
		"threshold_time":   thresholdTime.Format(time.RFC3339),
	}).Debug("Cleaning up old Spark results from data lake")

	// List all directories in the output path
	dirs, err := sbp.dataLakeClient.ListDirectories(ctx, sbp.outputPath)
	if err != nil {
		return 0, fmt.Errorf("failed to list directories: %w", err)
	}

	deletedCount := 0
	for _, dir := range dirs {
		// Get directory metadata (modification time)
		metadata, err := sbp.dataLakeClient.GetMetadata(ctx, dir)
		if err != nil {
			sbp.logger.WithError(err).WithField("directory", dir).Warn("Failed to get directory metadata")
			continue
		}

		// Check if directory is older than threshold
		if metadata.ModTime.Before(thresholdTime) {
			sbp.logger.WithFields(logrus.Fields{
				"directory": dir,
				"mod_time":  metadata.ModTime.Format(time.RFC3339),
				"threshold": thresholdTime.Format(time.RFC3339),
				"age_hours": time.Since(metadata.ModTime).Hours(),
			}).Debug("Deleting old directory")

			// Delete the directory
			if err := sbp.dataLakeClient.DeletePath(ctx, dir, true); err != nil {
				sbp.logger.WithError(err).WithField("directory", dir).Error("Failed to delete directory")
				continue
			}

			deletedCount++
			sbp.logger.WithField("directory", dir).Info("Deleted old Spark results directory")
		}
	}

	sbp.logger.WithField("deleted_count", deletedCount).Info("Cleanup of old Spark results completed")
	return deletedCount, nil
}
