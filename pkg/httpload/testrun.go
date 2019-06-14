package httpload

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

var startTime = time.Now()

func (s *Stage) runtest(resultChan chan *Result) {
	for ii := 1; ii <= s.Duration; ii++ {
		var wg sync.WaitGroup
		wg.Add(s.Concurrent)
		startTime = startTime.Add(time.Second)
		_rc := make(chan *Result, s.Concurrent)
		for i := 0; i < s.Concurrent; i++ {
			r := &Result{
				//StartTime:  time.Now(),
				//StartTime:  startTime,
				Concurrent: s.Concurrent,
			}
			rand.Seed(time.Now().UnixNano())
			responseTime := float64(rand.Intn(999)) / 1000
			fmt.Sprintf("responseTime:%f ", responseTime)
			if ii < 4 {
				responseTime = responseTime + float64(3)
			}
			go func() {
				defer wg.Done()
				s.doTest(r, responseTime, _rc)
			}()
			time.Sleep(time.Microsecond)
		}
		wg.Wait()
		close(_rc)

		// TODO 对当前秒求平均值
		var responseTimeTotal float64
		var podNumTotal float64
		for result := range _rc {
			responseTimeTotal += result.ResponseTime
			podNumTotal += result.PodNum
		}

		resultChan <- &Result{
			Concurrent:   s.Concurrent,
			PodNum:       podNumTotal / float64(s.Concurrent),
			ResponseTime: responseTimeTotal / float64(s.Concurrent),
		}
	}
}

func (s *Stage) doTest(result *Result, responseTime float64, resultChan chan *Result) {
	podNumChan := make(chan float64, 1)
	go func(pm chan float64) {
		// TODO get pod Num
		//pm <- rand.Intn(result.Concurrent)
		pm <- math.Ceil(float64(result.Concurrent) / 3)

	}(podNumChan)

	// TODO do request
	if responseTime != 0.0 {
		//result.EndTime = result.StartTime.Add(time.Duration(float64(time.Second) * responseTime))
	} else {
		//r := rand.Intn(5000)
		//timelength := time.Microsecond * time.Duration(r)
		//result.EndTime = time.Now().Add(timelength)
	}
	result.ResponseTime = responseTime

	result.PodNum = <-podNumChan

	resultChan <- result
	// TODO log result
	//result.log()
}
