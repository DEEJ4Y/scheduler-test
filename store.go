package main

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/DEEJ4Y/scheduler"
)

// InMemoryStore is a thread-safe, in-memory implementation of the JobStore interface.
// It uses a slice to store jobs and a mutex to ensure concurrent access safety.
type InMemoryStore struct {
	mu   sync.Mutex
	jobs []*scheduler.Job
}

// NewInMemoryStore creates a new instance of InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		jobs: make([]*scheduler.Job, 0),
	}
}

// AddJob adds a new job to the store. This is a helper method for testing.
func (s *InMemoryStore) AddJob(job *scheduler.Job) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, job)
}

// LockNext finds and atomically locks the next available job for processing.
// It returns the job BEFORE locking it (with original sleepUntil value).
// A job is available if:
//   - sleepUntil is not nil
//   - sleepUntil is <= current time
//
// The method locks the job by setting its sleepUntil to lockUntil.
func (s *InMemoryStore) LockNext(ctx context.Context, lockUntil time.Time) (*scheduler.Job, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Find the first available job
	for _, job := range s.jobs {
		// Skip jobs that are completed (sleepUntil is nil)
		if job.SleepUntil == nil {
			continue
		}

		// Check if the job is ready to be processed
		if job.SleepUntil.Before(now) || job.SleepUntil.Equal(now) {
			// Create a copy of the job with the original sleepUntil
			jobCopy := &scheduler.Job{
				ID:          job.ID,
				SleepUntil:  job.SleepUntil,
				Interval:    job.Interval,
				RepeatUntil: job.RepeatUntil,
				AutoRemove:  job.AutoRemove,
				Data:        job.Data,
			}

			// Lock the job by updating its sleepUntil
			job.SleepUntil = &lockUntil

			return jobCopy, nil
		}
	}

	// No available jobs
	return nil, nil
}

// Update modifies the fields of a job identified by jobID.
// It applies the updates specified in the JobUpdate struct.
func (s *InMemoryStore) Update(ctx context.Context, jobID interface{}, updates scheduler.JobUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the job by ID
	for _, job := range s.jobs {
		if job.ID == jobID {
			// Apply updates - JobUpdate only has SleepUntil field
			if updates.SleepUntil != nil {
				job.SleepUntil = *updates.SleepUntil
			}
			return nil
		}
	}

	return errors.New("job not found")
}

// Remove deletes a job from the store by its ID.
// This is typically used for auto-remove jobs after they complete.
func (s *InMemoryStore) Remove(ctx context.Context, jobID interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and remove the job
	for i, job := range s.jobs {
		if job.ID == jobID {
			// Remove by creating a new slice without this job
			s.jobs = append(s.jobs[:i], s.jobs[i+1:]...)
			return nil
		}
	}

	return errors.New("job not found")
}

// GetJobCount returns the current number of jobs in the store.
// This is a helper method for testing and monitoring.
func (s *InMemoryStore) GetJobCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.jobs)
}
