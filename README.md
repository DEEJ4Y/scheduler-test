# Scheduler Test Repository

This repository contains test implementations for the [github.com/DEEJ4Y/scheduler](https://github.com/DEEJ4Y/scheduler) package.

## Overview

Two test implementations are provided:
1. **In-Memory Store** - Simple, self-contained test using an in-memory job store
2. **MongoDB Store** - Production-ready test using MongoDB for job persistence

## Quick Start

### In-Memory Test

No dependencies required. Just run:

```bash
go run . -test=inmemory
```

or

```bash
go run . -test=mem
```

### MongoDB Test

**Prerequisites:**
- MongoDB running locally on `localhost:27017`

Start MongoDB (example):
```bash
# Using Docker
docker run -d -p 27017:27017 --name mongodb mongo:latest

# Or using system MongoDB
sudo systemctl start mongod   # Linux
brew services start mongodb   # macOS
```

Run the test:
```bash
go run . -test=mongodb
```

or

```bash
go run . -test=mongo
```

## Test Details

Both tests perform identical operations:
- Create 5 one-time jobs with different delays:
  - Job 1: Immediate execution (0s)
  - Job 2: Execute after 2 seconds
  - Job 3: Execute after 5 seconds
  - Job 4: Execute after 8 seconds
  - Job 5: Execute after 10 seconds
- Run scheduler for 15 seconds
- Auto-remove jobs after execution
- Graceful shutdown with statistics

### Scheduler Configuration

Both tests use identical scheduler settings:
- **NextDelay**: 100ms (delay between job checks)
- **IdleDelay**: 500ms (delay when no jobs available)
- **LockDuration**: 1 minute (job lock timeout)

### Event Handlers

All tests implement comprehensive event handlers:
- `OnDocument`: Logs job processing with timestamp and data
- `OnError`: Logs any errors that occur
- `OnIdle`: Logs when scheduler is idle (no jobs available)
- `OnStart`: Logs scheduler startup
- `OnStop`: Logs scheduler shutdown

## File Structure

```
.
├── main.go           # Test runner with CLI
├── inmemory.go       # In-memory store test
├── mongodb.go        # MongoDB store test
├── store.go          # InMemoryStore implementation
├── go.mod            # Go module definition
└── README.md         # This file
```

## InMemoryStore Implementation

The `store.go` file contains a complete implementation of the `scheduler.JobStore` interface:

```go
type JobStore interface {
    LockNext(ctx context.Context, lockUntil time.Time) (*Job, error)
    Update(ctx context.Context, jobID interface{}, updates JobUpdate) error
    Remove(ctx context.Context, jobID interface{}) error
}
```

Features:
- Thread-safe using `sync.Mutex`
- Stores jobs in a slice
- Proper time-based job availability checking
- Atomic job locking

## MongoDB Test Details

The MongoDB test:
- Connects to `mongodb://localhost:27017`
- Uses database: `scheduler_test`
- Uses collection: `jobs`
- Clears existing jobs before each run
- Shows MongoDB document IDs in logs

## Example Output

```
Scheduler Test Runner
=====================
Running test: inmemory

========================================
  In-Memory Scheduler Test Application
========================================

[INIT] Created in-memory job store
[INIT] Scheduler configured successfully
[INIT] Settings: NextDelay=100ms, IdleDelay=500ms, LockDuration=1m

[QUEUE] Adding jobs to the queue...
[16:29:08.407] [QUEUE] Added: Job 1 (delay: 0s, scheduled: 16:29:08.407)
[16:29:08.407] [QUEUE] Added: Job 2 (delay: 2s, scheduled: 16:29:10.407)
[16:29:08.407] [QUEUE] Added: Job 3 (delay: 5s, scheduled: 16:29:13.407)
[16:29:08.407] [QUEUE] Added: Job 4 (delay: 8s, scheduled: 16:29:16.407)
[16:29:08.407] [QUEUE] Added: Job 5 (delay: 10s, scheduled: 16:29:18.407)
[QUEUE] Total jobs queued: 5

========================================
    Starting Scheduler
========================================

[16:29:08.407] [START] Scheduler started
[INFO] Scheduler will run for 15 seconds...

[16:29:08.508] [PROCESSING] Job: Job 1 | Task: immediate execution | Data: map[name:Job 1 task:immediate execution]
[16:29:08.608] [IDLE] No jobs available, waiting...
[16:29:10.413] [PROCESSING] Job: Job 2 | Task: execute after 2 seconds | Data: map[name:Job 2 task:execute after 2 seconds]
...

========================================
    Final Statistics
========================================
Jobs remaining in store: 0
Scheduler is idle: true
Scheduler is processing: false

[DONE] In-Memory test completed successfully!
```

## Dependencies

```
go 1.24.7

require (
    github.com/DEEJ4Y/scheduler v1.0.0
    github.com/DEEJ4Y/scheduler/mongodb v1.0.0  // For MongoDB test
    go.mongodb.org/mongo-driver v1.17.4          // For MongoDB test
)
```

## Usage Tips

1. **Default test**: Running without flags defaults to in-memory test:
   ```bash
   go run .
   ```

2. **Help**: See available options:
   ```bash
   go run . -test=help
   ```

3. **Watch MongoDB**: To monitor jobs in MongoDB during test:
   ```bash
   # In another terminal
   mongosh
   use scheduler_test
   db.jobs.find().pretty()
   ```

## Notes

- Both tests complete in approximately 15 seconds
- Jobs are auto-removed after execution
- All jobs should execute exactly once
- Final statistics show 0 remaining jobs
- MongoDB test requires MongoDB server running
- In-memory test has no external dependencies

## License

MIT License - see main scheduler package for details.
