package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/DEEJ4Y/scheduler"
	"github.com/DEEJ4Y/scheduler/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func runMongoDBTest() {
	fmt.Println("========================================")
	fmt.Println("  MongoDB Scheduler Test Application")
	fmt.Println("========================================")
	fmt.Println()

	// Create context for MongoDB connection
	ctx := context.Background()

	// Connect to MongoDB (standard local host and port)
	mongoURI := "mongodb://localhost:27017"
	fmt.Printf("[INIT] Connecting to MongoDB at %s\n", mongoURI)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("[FATAL] Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		fmt.Println("[CLEANUP] Disconnecting from MongoDB...")
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("[ERROR] Error disconnecting from MongoDB: %v", err)
		}
	}()

	// Ping MongoDB to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("[FATAL] Failed to ping MongoDB: %v", err)
	}
	fmt.Println("[INIT] Successfully connected to MongoDB")

	// Get collection for jobs
	database := client.Database("scheduler_test")
	collection := database.Collection("jobs")

	// Clear any existing jobs from previous runs
	fmt.Println("[CLEANUP] Clearing existing jobs from collection...")
	if _, err := collection.DeleteMany(ctx, bson.M{}); err != nil {
		log.Printf("[WARN] Failed to clear collection: %v", err)
	}

	// Create MongoDB store
	store, err := mongodb.NewStore(mongodb.Config{
		Collection: collection,
	})
	if err != nil {
		log.Fatalf("[FATAL] Failed to create MongoDB store: %v", err)
	}
	fmt.Println("[INIT] Created MongoDB job store")

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
	for _, j := range jobs {
		sleepUntil := now.Add(j.delay)
		job := bson.M{
			"sleepUntil": sleepUntil,
			"autoRemove": true, // Automatically remove after execution
			"name":       j.name,
			"task":       j.task,
		}

		result, err := collection.InsertOne(ctx, job)
		if err != nil {
			log.Printf("[ERROR] Failed to insert job: %v", err)
			continue
		}

		timestamp := time.Now().Format("15:04:05.000")
		fmt.Printf("[%s] [QUEUE] Added: %s (delay: %v, scheduled: %s, ID: %v)\n",
			timestamp, j.name, j.delay, sleepUntil.Format("15:04:05.000"), result.InsertedID)
	}

	// Count jobs in collection
	jobCount, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("[WARN] Failed to count jobs: %v", err)
	} else {
		fmt.Printf("[QUEUE] Total jobs queued: %d\n", jobCount)
	}
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

	jobCount, err = collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Printf("[ERROR] Failed to count remaining jobs: %v", err)
	} else {
		fmt.Printf("Jobs remaining in MongoDB: %d\n", jobCount)
	}

	fmt.Printf("Scheduler is idle: %v\n", sched.IsIdle())
	fmt.Printf("Scheduler is processing: %v\n", sched.IsProcessing())
	fmt.Println()
	fmt.Println("[DONE] MongoDB test completed successfully!")
}
