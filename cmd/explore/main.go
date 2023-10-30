package main

import (
	"time"

	"github.com/MBvisti/grafto/pkg/telemetry"
	"github.com/robfig/cron/v3"
)

// func watcher(jobs chan database.Job) {
// 	for i := 1; i < 1000; i++ {
// 		jobs <- database.Job{
// 			ID:        uuid.New(),
// 			CreatedAt: database.ConvertTime(time.Now()),
// 			UpdatedAt: database.ConvertTime(time.Now()),
// 		}
// 	}
// }

// func process(wg *sync.WaitGroup, job database.Job) {
// 	telemetry.Logger.Info("processing job", "job", job.ID)
// 	time.Sleep(time.Second)

// 	wg.Done()
// }

// func worker(id int, jobs chan database.Job) {
// 	var wg sync.WaitGroup

// 	for job := range jobs {
// 		wg.Add(1)
// 		telemetry.Logger.Info("worker received jobs", "id", id, "count", job)

// 		telemetry.Logger.Info("processing job", "job", job.ID)
// 		time.Sleep(time.Second)

// 		wg.Done()
// 		// go process(&wg, job)
// 	}

// 	wg.Wait()
// }

func main() {
	// jobsStream := make(chan database.Job)
	// sema := make(chan struct{}, 2)

	// go watcher(jobsStream)

	// for i := 1; i <= 5; i++ {
	// 	sema <- struct{}{}
	// 	go worker(i, jobsStream)
	// }

	// sigCh := make(chan os.Signal, 1)
	// signal.Notify(sigCh, syscall.SIGTERM)

	// // The code blocks until a signal is received (e.g. Ctrl+C).
	// <-sigCh

	// close(jobsStream)

	// timeFormat := "2006/01/02 03:04:05 UTC"

	// st := time.Now()
	// startTime := time.Date(st.Year(), st.Month(), st.Day(), st.Hour(), st.Minute(), 00, 00, time.UTC)
	// now, err := gronx.NextTickAfter("* * * * *", startTime, true)
	// if err != nil {
	// 	panic(err)
	// }
	sched := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	parser, err := sched.Parse("* * * * *")
	if err != nil {
		panic(err)
	}

	iter := 1
	for {
		// t := time.Now()
		// tickTime := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 00, 00, time.UTC)
		// telemetry.Logger.Info("tick time is", "time", tickTime)
		// nextTick, err := gronx.NextTick("* * * * *", true)
		// if err != nil {
		// 	panic(err)
		// }

		telemetry.Logger.Info("next tick is", "time", parser.Next(time.Now()))

		iter++
		time.Sleep(125 * time.Millisecond)
	}
}
