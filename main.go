package main

import (
	"fmt"
	"math"
	"time"
)

const NUM_WORKERS = 3
const TOTAL_WORK = 12 * time.Second
const EXPECTED = float64(TOTAL_WORK / NUM_WORKERS)

func main() {
	work_queue := make(chan time.Duration)
	result_queue := make(chan time.Duration)
	done_consuming := make(chan bool)
	death_queue := make(chan bool)

	// start the workers
	for i := 0; i < NUM_WORKERS; i++ {
		go Worker(work_queue, result_queue, death_queue)
		// ensure its ultimate demise
		defer kill(death_queue)
	}

	for e := 2; e < 7; e++ {
		// start a consumer who will let us know when the work is all done
		go Consumer(result_queue, TOTAL_WORK, done_consuming)

		job_size := time.Duration(math.Pow10(e)) * time.Microsecond
		fmt.Println("job size: ", job_size)

		// shove jobs into the work queue
		start := time.Now()
		for j := TOTAL_WORK; j > time.Duration(0); j -= job_size {
			work_queue <- job_size
		}

		// wait until the consumer is done
		<-done_consuming

		// report findings
		elapsed := time.Now().Sub(start)
		fmt.Println("done in ", elapsed)
		penalty := (float64(elapsed) - EXPECTED) / EXPECTED
		fmt.Printf("overhead: %0.2f%%\n", penalty*100)
		fmt.Println("")
	}
}

// sends "true" into the given channel
func kill(death_queue chan bool) {
	death_queue <- true
}

// counts down the total amount of work until it's all been accounted for, then
// sends "true" on the "done" channel.
func Consumer(results chan time.Duration, total_work time.Duration, done chan bool) {
	for total_work > 0 {
		total_work -= <-results
	}
	// notify whoever cares that we're all done here
	done <- true
}

// Accepts jobs on the "in" channel and sleeps for the specified duration,
// which simulates doing that much work (reminds me of a former colleague).
// When done, sends the job's duration down the "out" channel. Returns when any
// value is received on the "die" channel.
func Worker(in chan time.Duration, out chan time.Duration, die chan bool) {
	for {
		select {
		case interval := <-in:
			time.Sleep(interval)
			out <- interval
		case <-die:
			return
		}
	}
}
