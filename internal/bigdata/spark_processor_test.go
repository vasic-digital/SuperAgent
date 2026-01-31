package bigdata

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newTestSparkProcessor() *SparkBatchProcessor {
	return &SparkBatchProcessor{
		logger: logrus.New(),
	}
}

func TestParseJobOutput_EntityExtraction(t *testing.T) {
	processor := newTestSparkProcessor()
	params := BatchParams{
		JobType: BatchJobEntityExtraction,
	}

	output := `{"processed_rows": 150000, "entities_extracted": 75000, "status": "completed"}`

	result, err := processor.parseJobOutput(output, params)
	assert.NoError(t, err)
	assert.Equal(t, BatchJobEntityExtraction, result.JobType)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, int64(150000), result.ProcessedRows)
	assert.Equal(t, int64(75000), result.EntitiesExtracted)
}

func TestParseJobOutput_RelationshipMining(t *testing.T) {
	processor := newTestSparkProcessor()
	params := BatchParams{
		JobType: BatchJobRelationshipMining,
	}

	output := `{"processed_rows": 120000, "relationships_found": 30000, "status": "completed"}`

	result, err := processor.parseJobOutput(output, params)
	assert.NoError(t, err)
	assert.Equal(t, BatchJobRelationshipMining, result.JobType)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, int64(120000), result.ProcessedRows)
	assert.Equal(t, int64(30000), result.RelationshipsFound)
}

func TestParseJobOutput_InvalidJSON(t *testing.T) {
	processor := newTestSparkProcessor()
	params := BatchParams{
		JobType: BatchJobEntityExtraction,
	}

	output := `invalid json`

	result, err := processor.parseJobOutput(output, params)
	assert.Error(t, err) // Should return error when no valid JSON found
	assert.Nil(t, result)
}

func TestParseJobOutput_EmptyOutput(t *testing.T) {
	processor := newTestSparkProcessor()
	params := BatchParams{
		JobType: BatchJobEntityExtraction,
	}

	output := ""

	result, err := processor.parseJobOutput(output, params)
	assert.Error(t, err)
	assert.Nil(t, result)
}
