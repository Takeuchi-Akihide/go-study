package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type JobStatus string

const (
	JobQueued  JobStatus = "queued"
	JobRunning JobStatus = "running"
	JobDone    JobStatus = "done"
	JobFailed  JobStatus = "failed"
)

const (
	jobEnqueueTimeout = 500 * time.Millisecond
	jobProcessTimeout = 3 * time.Second
)

type Job struct {
	ID        int       `json:"id"`
	Status    JobStatus `json:"status"`
	Input     string    `json:"input"`
	Result    string    `json:"result,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type JobRequest struct {
	Input string `json:"input"`
}

var (
	jobQueue = make(chan int, 10)

	jobsMu    sync.Mutex
	jobs      = map[int]*Job{}
	nextJobID = 1
)

func createJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req JobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	now := time.Now()
	jobsMu.Lock()
	id := nextJobID
	nextJobID++

	job := &Job{
		ID:        id,
		Status:    JobQueued,
		Input:     req.Input,
		CreatedAt: now,
		UpdatedAt: now,
	}
	jobs[id] = job
	jobsMu.Unlock()

	ctx, cancel := context.WithTimeout(r.Context(), jobEnqueueTimeout)
	defer cancel()

	select {
	case jobQueue <- id:
		writeJSON(w, http.StatusAccepted, job)
	case <-ctx.Done():
		updateJobStatus(id, JobFailed, "job queue is busy")
		writeError(w, http.StatusServiceUnavailable, "job queue is busy")
	}
}

func getJobHandler(w http.ResponseWriter, r *http.Request) {
	idText := r.URL.Path[len("/jobs/"):]

	id, err := strconv.Atoi(idText)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid job id")
		return
	}

	jobsMu.Lock()
	job, ok := jobs[id]
	jobsMu.Unlock()

	if !ok {
		writeError(w, http.StatusNotFound, "job not found")
		return
	}

	writeJSON(w, http.StatusOK, job)
}

func startWorkers(n int) {
	for i := 0; i < n; i++ {
		go worker(i + 1)
	}
}

func worker(workerID int) {
	for jobID := range jobQueue {
		ctx, cancel := context.WithTimeout(context.Background(), jobProcessTimeout)

		if err := processJob(ctx, workerID, jobID); err != nil {
			updateJobStatus(jobID, JobFailed, err.Error())
		}

		cancel()
	}
}

func processJob(ctx context.Context, workerID int, jobID int) error {
	updateJobStatus(jobID, JobRunning, "")

	select {
	case <-time.After(2 * time.Second):
		result := "processed by worker " + strconv.Itoa(workerID)
		updateJobStatus(jobID, JobDone, result)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}

}

func updateJobStatus(id int, status JobStatus, result string) {
	jobsMu.Lock()
	defer jobsMu.Unlock()

	job, ok := jobs[id]
	if !ok {
		return
	}

	job.Status = status
	if result != "" {
		job.Result = result
	}
	job.UpdatedAt = time.Now()
}
