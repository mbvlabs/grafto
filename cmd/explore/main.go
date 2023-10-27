package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/MBvisti/grafto/repository/database"
	"github.com/google/uuid"
)

func watcher(jobs chan database.Job) {
	for i := 1; i < 1000; i++ {
		jobs <- database.Job{
			ID:        uuid.New(),
			CreatedAt: database.ConvertTime(time.Now()),
			UpdatedAt: database.ConvertTime(time.Now()),
		}
	}
}

// func process(wg *sync.WaitGroup, job database.Job) {
// 	telemetry.Logger.Info("processing job", "job", job.ID)
// 	time.Sleep(time.Second)

// 	wg.Done()
// }

func worker(id int, jobs chan database.Job) {
	var wg sync.WaitGroup

	for job := range jobs {
		wg.Add(1)
		telemetry.Logger.Info("worker received jobs", "id", id, "count", job)

		telemetry.Logger.Info("processing job", "job", job.ID)
		time.Sleep(time.Second)

		wg.Done()
		// go process(&wg, job)
	}

	wg.Wait()
}

func main() {
	jobsStream := make(chan database.Job)

	go watcher(jobsStream)

	for i := 1; i <= 5; i++ {
		go worker(i, jobsStream)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM)

	// The code blocks until a signal is received (e.g. Ctrl+C).
	<-sigCh

	close(jobsStream)
}
