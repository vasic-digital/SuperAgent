package flink

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		client, err := NewClient(nil, nil)
		require.NoError(t, err)
		assert.NotNil(t, client)
		assert.False(t, client.IsConnected())
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &Config{
			JobManagerHost:     "flink-server",
			JobManagerPort:     6123,
			WebUIPort:          8082,
			RESTURL:            "http://flink-server:8082",
			RequestTimeout:     60 * time.Second,
			DefaultParallelism: 4,
		}
		client, err := NewClient(config, logrus.New())
		require.NoError(t, err)
		assert.NotNil(t, client)
	})

	t.Run("with invalid config", func(t *testing.T) {
		config := &Config{
			JobManagerHost: "",
		}
		client, err := NewClient(config, nil)
		require.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestClientConnect(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/overview" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"taskmanagers":    2,
					"slots-total":     8,
					"slots-available": 4,
					"jobs-running":    1,
					"flink-version":   "1.18.0",
				})
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		config := DefaultConfig()
		config.RESTURL = server.URL
		client, err := NewClient(config, nil)
		require.NoError(t, err)

		err = client.Connect(context.Background())
		require.NoError(t, err)
		assert.True(t, client.IsConnected())
	})

	t.Run("connection failure", func(t *testing.T) {
		config := DefaultConfig()
		config.RESTURL = "http://localhost:99999"
		config.RequestTimeout = 100 * time.Millisecond
		client, err := NewClient(config, nil)
		require.NoError(t, err)

		err = client.Connect(context.Background())
		require.Error(t, err)
		assert.False(t, client.IsConnected())
	})
}

func TestClientClose(t *testing.T) {
	client, _ := NewClient(nil, nil)
	err := client.Close()
	require.NoError(t, err)
	assert.False(t, client.IsConnected())
}

func TestClientHealthCheck(t *testing.T) {
	t.Run("healthy cluster", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"taskmanagers":  2,
				"flink-version": "1.18.0",
			})
		}))
		defer server.Close()

		config := DefaultConfig()
		config.RESTURL = server.URL
		client, _ := NewClient(config, nil)
		client.Connect(context.Background())

		err := client.HealthCheck(context.Background())
		require.NoError(t, err)
	})
}

func TestGetOverview(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"taskmanagers":    4,
				"slots-total":     16,
				"slots-available": 8,
				"jobs-running":    2,
				"jobs-finished":   10,
				"jobs-cancelled":  1,
				"jobs-failed":     0,
				"flink-version":   "1.18.0",
				"flink-commit":    "abc123",
			})
		}))
		defer server.Close()

		config := DefaultConfig()
		config.RESTURL = server.URL
		client, _ := NewClient(config, nil)
		client.Connect(context.Background())

		overview, err := client.GetOverview(context.Background())
		require.NoError(t, err)
		assert.Equal(t, 4, overview.TaskManagers)
		assert.Equal(t, 16, overview.SlotsTotal)
		assert.Equal(t, 8, overview.SlotsAvailable)
		assert.Equal(t, 2, overview.JobsRunning)
		assert.Equal(t, "1.18.0", overview.FlinkVersion)
	})
}

func TestGetTaskManagers(t *testing.T) {
	t.Run("successful listing", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/taskmanagers" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"taskmanagers": []map[string]interface{}{
						{
							"id":          "tm-001",
							"path":        "akka://flink@host:6122",
							"dataPort":    6121,
							"slotsNumber": 4,
							"freeSlots":   2,
							"hardware": map[string]interface{}{
								"cpuCores":       8,
								"physicalMemory": 17179869184,
								"managedMemory":  8589934592,
							},
						},
					},
				})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"taskmanagers": 1})
		}))
		defer server.Close()

		config := DefaultConfig()
		config.RESTURL = server.URL
		client, _ := NewClient(config, nil)
		client.Connect(context.Background())

		tms, err := client.GetTaskManagers(context.Background())
		require.NoError(t, err)
		assert.Len(t, tms, 1)
		assert.Equal(t, "tm-001", tms[0].ID)
		assert.Equal(t, 4, tms[0].SlotsNumber)
	})
}

func TestGetJobs(t *testing.T) {
	t.Run("successful listing", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/jobs" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"jobs": []map[string]interface{}{
						{
							"id":         "job-001",
							"name":       "TestJob",
							"state":      "RUNNING",
							"start-time": 1704067200000,
							"end-time":   -1,
							"duration":   3600000,
						},
					},
				})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"taskmanagers": 1})
		}))
		defer server.Close()

		config := DefaultConfig()
		config.RESTURL = server.URL
		client, _ := NewClient(config, nil)
		client.Connect(context.Background())

		jobs, err := client.GetJobs(context.Background())
		require.NoError(t, err)
		assert.Len(t, jobs, 1)
		assert.Equal(t, "job-001", jobs[0].ID)
		assert.Equal(t, "RUNNING", jobs[0].State)
	})
}

func TestGetJob(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/jobs/job-001" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"jid":        "job-001",
					"name":       "TestJob",
					"state":      "RUNNING",
					"start-time": 1704067200000,
					"duration":   3600000,
				})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"taskmanagers": 1})
		}))
		defer server.Close()

		config := DefaultConfig()
		config.RESTURL = server.URL
		client, _ := NewClient(config, nil)
		client.Connect(context.Background())

		job, err := client.GetJob(context.Background(), "job-001")
		require.NoError(t, err)
		assert.Equal(t, "job-001", job.JID)
		assert.Equal(t, "RUNNING", job.State)
	})
}

func TestCancelJob(t *testing.T) {
	t.Run("successful cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/jobs/job-001" && r.Method == http.MethodPatch {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"taskmanagers": 1})
		}))
		defer server.Close()

		config := DefaultConfig()
		config.RESTURL = server.URL
		client, _ := NewClient(config, nil)
		client.Connect(context.Background())

		err := client.CancelJob(context.Background(), "job-001")
		require.NoError(t, err)
	})
}

func TestGetJars(t *testing.T) {
	t.Run("successful listing", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/jars" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"files": []map[string]interface{}{
						{"id": "jar-001", "name": "job.jar"},
						{"id": "jar-002", "name": "analytics.jar"},
					},
				})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"taskmanagers": 1})
		}))
		defer server.Close()

		config := DefaultConfig()
		config.RESTURL = server.URL
		client, _ := NewClient(config, nil)
		client.Connect(context.Background())

		jars, err := client.GetJars(context.Background())
		require.NoError(t, err)
		assert.Len(t, jars, 2)
	})
}

func TestDeleteJar(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/jars/jar-001" && r.Method == http.MethodDelete {
				w.WriteHeader(http.StatusOK)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"taskmanagers": 1})
		}))
		defer server.Close()

		config := DefaultConfig()
		config.RESTURL = server.URL
		client, _ := NewClient(config, nil)
		client.Connect(context.Background())

		err := client.DeleteJar(context.Background(), "jar-001")
		require.NoError(t, err)
	})
}

func TestSubmitJob(t *testing.T) {
	t.Run("successful submission", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/jars/jar-001/run" && r.Method == http.MethodPost {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"jobid": "new-job-001",
				})
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{"taskmanagers": 1})
		}))
		defer server.Close()

		config := DefaultConfig()
		config.RESTURL = server.URL
		client, _ := NewClient(config, nil)
		client.Connect(context.Background())

		jobConfig := &JobConfig{
			Parallelism: 4,
		}
		jobID, err := client.SubmitJob(context.Background(), "jar-001", jobConfig)
		require.NoError(t, err)
		assert.Equal(t, "new-job-001", jobID)
	})
}

func TestClusterOverviewType(t *testing.T) {
	overview := &ClusterOverview{
		FlinkVersion:   "1.18.0",
		FlinkCommit:    "abc123",
		TaskManagers:   4,
		SlotsTotal:     16,
		SlotsAvailable: 8,
		JobsRunning:    2,
		JobsFinished:   10,
		JobsCancelled:  1,
		JobsFailed:     0,
	}

	assert.Equal(t, "1.18.0", overview.FlinkVersion)
	assert.Equal(t, 4, overview.TaskManagers)
	assert.Equal(t, 16, overview.SlotsTotal)
}

func TestTaskManagerType(t *testing.T) {
	tm := TaskManager{
		ID:          "tm-001",
		Path:        "akka://flink@host:6122",
		DataPort:    6121,
		SlotsNumber: 4,
		FreeSlots:   2,
	}

	assert.Equal(t, "tm-001", tm.ID)
	assert.Equal(t, 4, tm.SlotsNumber)
}

func TestJobType(t *testing.T) {
	job := Job{
		ID:        "job-001",
		Name:      "TestJob",
		State:     "RUNNING",
		StartTime: 1704067200000,
		EndTime:   -1,
		Duration:  3600000,
	}

	assert.Equal(t, "job-001", job.ID)
	assert.Equal(t, "TestJob", job.Name)
	assert.Equal(t, "RUNNING", job.State)
}

func TestJobDetailsType(t *testing.T) {
	details := JobDetails{
		JID:         "job-001",
		Name:        "TestJob",
		IsStoppable: true,
		State:       "RUNNING",
		StartTime:   1704067200000,
		Duration:    3600000,
	}

	assert.Equal(t, "job-001", details.JID)
	assert.Equal(t, "TestJob", details.Name)
	assert.True(t, details.IsStoppable)
}
