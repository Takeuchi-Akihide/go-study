package main

import (
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

	jobs[id] = &Job{
		ID:        id,
		Status:    JobQueued,
		Input:     req.Input,
		CreatedAt: now,
		UpdatedAt: now,
	}
	jobsMu.Unlock()

	select {
	case jobQueue <- id:
		writeJSON(w, http.StatusAccepted, jobs[id])
	default:
		writeError(w, http.StatusServiceUnavailable, "job queue is full")
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
		updateJobStatus(jobID, JobRunning, "")

		time.Sleep(20 * time.Second)

		result := "processed by worker" + strconv.Itoa(workerID)

		updateJobStatus(jobID, JobDone, result)
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
