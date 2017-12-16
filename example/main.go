package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/penhauer-xiao/tunny"
)

func closure() {
	fmt.Println("begin ...closure")
	exampleChannel := make(chan int)

	numCPUs := runtime.NumCPU()
	pool, _ := tunny.CreatePoolGeneric(numCPUs).Open()

	_, err := pool.SendWork(func() {
		/* Do your hard work here, usual rules of closures apply here,
		 * so you can return values like so:
		 */
		fmt.Println("closure work...")
		exampleChannel <- 10
	})

	if err != nil {
		// You done goofed
	}

	defer pool.Close()
}

// ++++++++++++++++++++++++++++++++++++
type customWorker struct {
	// TODO: Put some state here
}

// Use this call to block further jobs if necessary
func (worker *customWorker) TunnyReady() bool {
	return true
}

func (worker *customWorker) TunnyInitialize() {
	fmt.Println("init...")
}

func (worker *customWorker) TunnyTerminate() {
	fmt.Println("Terminate...")
}

func (worker *customWorker) TunnyInterrupt() {
	fmt.Println("Interrupt...")
}

// This is where the work actually happens
func (worker *customWorker) TunnyJob(data interface{}) interface{} {
	/* TODO: Use and modify state
	 * there's no need for thread safety paradigms here unless the
	 * data is being accessed from another goroutine outside of
	 * the pool.
	 */
	if outputStr, ok := data.(string); ok {
		if outputStr == "ko" {
			time.Sleep(time.Second * 70)
		}
		if outputStr == "ok" {
			time.Sleep(time.Second * 200)
		}
		return ("custom job done: " + outputStr)
	}
	return nil
}

func TestCustomWorkers() {
	fmt.Println("TestCustomWorkers")
	// outChan:= make(chan int, 10)

	workers := make([]tunny.TunnyWorker, 4)
	for i, _ := range workers {
		workers[i] = &(customWorker{})
	}

	pool, _ := tunny.CreateCustomPool(workers).Open()
	defer pool.Close()

	for i := 0; i < 50; i++ {
		if i < 2 {
			value, _ := pool.SendWork("ko")
			fmt.Println(value.(string))
		} else if i < 4 {
			value, _ := pool.SendWork("ok")
			fmt.Println(value.(string))
		} else {
			value, _ := pool.SendWork("hello world")
			fmt.Println(value.(string))
		}
	}
}

func Async() {
	fmt.Println("begin ... Async")
	// outChan:= make(chan int, 10)

	workers := make([]tunny.TunnyWorker, 4)
	for i, _ := range workers {
		workers[i] = &(customWorker{})
	}

	pool, _ := tunny.CreateCustomPool(workers).Open()
	defer pool.Close()

	for i := 0; i < 50; i++ {
		if i < 2 {
			pool.SendWorkAsync("ko", nil)
		} else if i < 4 {
			pool.SendWorkAsync("ok", nil)

		} else {
			pool.SendWorkAsync("hello world", nil)
		}
	}
	pool.PublishExpvarMetrics("KO")
	time.Sleep(time.Second * 100)
	fmt.Println("done ...Async")
}

func TestCustomWorkers1() {
	fmt.Println("TestCustomWorkers")
	// outChan:= make(chan int, 10)

	wg := new(sync.WaitGroup)
	wg.Add(2000)

	workers := make([]tunny.TunnyWorker, 4)
	for i, _ := range workers {
		workers[i] = &(customWorker{})
	}

	pool, _ := tunny.CreateCustomPool(workers).Open()

	defer pool.Close()

	for i := 0; i < 2000; i++ {
		go func() {
			if i < 2 {
				value, _ := pool.SendWork("ko")
				fmt.Println(value.(string))
			} else if i < 4 {
				value, _ := pool.SendWork("ok")
				fmt.Println(value.(string))
			} else {
				value, _ := pool.SendWork("hello world")
				fmt.Println(value.(string))
			}

			wg.Done()
		}()
	}

	wg.Wait()
}

// ++++++++++++++++++++++++++++++++++++

func main() {
	//Async()
	//return
	//closure()

	numCPUs := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPUs + 1) // numCPUs hot threads + one for async tasks.

	pool, _ := tunny.CreatePool(numCPUs, func(object interface{}) interface{} {
		input, _ := object.(string)

		// Do something that takes a lot of work
		output := input

		time.Sleep(time.Second * 5)

		fmt.Println("output:", output)

		return output
	}).Open()

	defer pool.Close()

	pool.PublishExpvarMetrics("KO")

	http.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}

		// Send work to our pool
		//result, _ := pool.SendWork(input)

		result, err := pool.SendWorkTimed(5000000000000, "xiaobenhua")
		if err != nil {
			http.Error(w, "Request timed out", http.StatusRequestTimeout)
		}

		w.Write([]byte(result.(string)))
	})

	http.ListenAndServe(":8080", nil)
}
