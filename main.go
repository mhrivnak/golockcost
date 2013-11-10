package main

import (
	"fmt"
	"math"
	"time"
)

const NUM_WORKERS = 3
const TOTAL_WORK = 12 * time.Second
const EXPECTED_SECONDS = float64(TOTAL_WORK / NUM_WORKERS)

func main() {
	work_queue := make(chan time.Duration)
	result_queue := make(chan time.Duration)
	done_receiving := make(chan bool)

	// start the workers
	for i := 0; i < NUM_WORKERS; i++ {
		go Worker(work_queue, result_queue)
	}

	for e := 2; e < 7; e++ {
		// start a receiver who will let us know when the work is all done
		go Receiver(result_queue, TOTAL_WORK, done_receiving)

		job_size := time.Duration(math.Pow10(e)) * time.Microsecond
		fmt.Println("job size: ", job_size)

		// shove jobs into the work queue
		start := time.Now()
		for j := TOTAL_WORK; j > time.Duration(0); j -= job_size {
			work_queue <- job_size
		}

		// wait until the receiver is done
		<-done_receiving

		// report findings
		elapsed := time.Now().Sub(start)
		fmt.Println("done in ", elapsed)
		penalty := (float64(elapsed) - EXPECTED_SECONDS) / EXPECTED_SECONDS
		fmt.Printf("overhead: %0.2f%%\n", penalty*100)
		fmt.Println("")
	}
}

// counts down the total amount of work until it's all been accounted for, then
// sends "true" on the "done" channel.
func Receiver(results chan time.Duration, total_work time.Duration, done chan bool) {
	for total_work > 0 {
		total_work -= <-results
	}
	// notify whoever cares that we're all done here
	done <- true
}

// Accepts jobs on the "in" channel and sleeps for the specified duration,
// which simulates doing that much work (reminds me of a former colleague).
// When done, sends the job's duration down the "out" channel.
func Worker(in chan time.Duration, out chan time.Duration) {
    var interval time.Duration
	for {
		interval = <-in
        time.Sleep(interval)
        out <- interval
	}
}
