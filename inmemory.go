package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DEEJ4Y/scheduler"
)

func runInMemoryTest() {
	fmt.Println("========================================")
	fmt.Println("  In-Memory Scheduler Test Application")
	fmt.Println("========================================")
	fmt.Println()

	// Create context for the scheduler
	ctx := context.Background()

	// Create an in-memory store
	store := NewInMemoryStore()
	fmt.Println("[INIT] Created in-memory job store")

	// Create and configure the scheduler
	sched, err := scheduler.New(scheduler.Config{
		Store:        store,
		NextDelay:    100 * time.Millisecond,
		IdleDelay:    500 * time.Millisecond,
		LockDuration: 1 * time.Minute,

		// OnDocument is called when a job is ready to be processed
		OnDocument: func(ctx context.Context, job *scheduler.Job) error {
			timestamp := time.Now().Format("15:04:05.000")
			jobName := "Unknown"
			jobTask := "Unknown"

			// Extract job data - job.Data is already map[string]interface{}
			if name, ok := job.Data["name"].(string); ok {
				jobName = name
			}
			if task, ok := job.Data["task"].(string); ok {
				jobTask = task
			}

			fmt.Printf("[%s] [PROCESSING] Job: %s | Task: %s | Data: %v\n",
				timestamp, jobName, jobTask, job.Data)

			return nil
		},

		// OnError is called when an error occurs
		OnError: func(ctx context.Context, err error) {
			timestamp := time.Now().Format("15:04:05.000")
			log.Printf("[%s] [ERROR] %v\n", timestamp, err)
		},

		// OnIdle is called when there are no jobs to process
		OnIdle: func(ctx context.Context) error {
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] [IDLE] No jobs available, waiting...\n", timestamp)
			return nil
		},

		// OnStart is called when the scheduler starts
		OnStart: func(ctx context.Context) error {
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] [START] Scheduler started\n", timestamp)
			return nil
		},

		// OnStop is called when the scheduler stops
		OnStop: func(ctx context.Context) error {
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] [STOP] Scheduler stopped\n", timestamp)
			return nil
		},
	})

	if err != nil {
		log.Fatalf("[FATAL] Failed to create scheduler: %v", err)
	}

	fmt.Println("[INIT] Scheduler configured successfully")
	fmt.Printf("[INIT] Settings: NextDelay=100ms, IdleDelay=500ms, LockDuration=1m\n")
	fmt.Println()

	// Add 5 one-time jobs with different delays
	now := time.Now()
	jobs := []struct {
		name  string
		delay time.Duration
		task  string
	}{
		{"Job 1", 0, "immediate execution"},
		{"Job 2", 2 * time.Second, "execute after 2 seconds"},
		{"Job 3", 5 * time.Second, "execute after 5 seconds"},
		{"Job 4", 8 * time.Second, "execute after 8 seconds"},
		{"Job 5", 10 * time.Second, "execute after 10 seconds"},
	}

	fmt.Println("[QUEUE] Adding jobs to the queue...")
	for i, j := range jobs {
		sleepUntil := now.Add(j.delay)
		job := &scheduler.Job{
			ID:         i + 1,
			SleepUntil: &sleepUntil,
			AutoRemove: true, // Automatically remove after execution
			Data: map[string]interface{}{
				"name": j.name,
				"task": j.task,
			},
		}
		store.AddJob(job)

		timestamp := time.Now().Format("15:04:05.000")
		fmt.Printf("[%s] [QUEUE] Added: %s (delay: %v, scheduled: %s)\n",
			timestamp, j.name, j.delay, sleepUntil.Format("15:04:05.000"))
	}

	fmt.Printf("[QUEUE] Total jobs queued: %d\n", store.GetJobCount())
	fmt.Println()

	// Start the scheduler
	fmt.Println("========================================")
	fmt.Println("    Starting Scheduler")
	fmt.Println("========================================")
	fmt.Println()

	sched.Start(ctx)

	// Run for 15 seconds
	fmt.Println("[INFO] Scheduler will run for 15 seconds...")
	fmt.Println()
	time.Sleep(15 * time.Second)

	// Graceful shutdown
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("    Shutting Down Scheduler")
	fmt.Println("========================================")
	fmt.Println()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = sched.Stop(shutdownCtx)
	if err != nil {
		log.Printf("[ERROR] Error during shutdown: %v", err)
	}

	// Final statistics
	fmt.Println()
	fmt.Println("========================================")
	fmt.Println("    Final Statistics")
	fmt.Println("========================================")
	fmt.Printf("Jobs remaining in store: %d\n", store.GetJobCount())
	fmt.Printf("Scheduler is idle: %v\n", sched.IsIdle())
	fmt.Printf("Scheduler is processing: %v\n", sched.IsProcessing())
	fmt.Println()
	fmt.Println("[DONE] In-Memory test completed successfully!")
}
