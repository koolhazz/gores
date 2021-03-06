package gores

import (
    "errors"
    "log"
    "sync"
    "time"
    "github.com/deckarep/golang-set"
)

var workerIdChan chan string

type Dispatcher struct {
    resq *ResQ
    maxWorkers int
    jobChannel chan *Job
    doneChannel chan int
    queues mapset.Set
    timeout int
}

func NewDispatcher(resq *ResQ, config *Config, queues mapset.Set) *Dispatcher{
    if resq == nil || config.MAX_WORKERS <= 0 {
        log.Println("Invalid number of workers to initialize Dispatcher")
        return nil
    }
    workerIdChan = make(chan string, config.MAX_WORKERS)
    return &Dispatcher{
              resq: resq,
              maxWorkers: config.MAX_WORKERS,
              jobChannel: make(chan *Job, config.MAX_WORKERS),
              queues: queues,
              timeout: config.DispatcherTimeout,
            }
}

func (disp *Dispatcher) Run(tasks *map[string]interface{}) error {
    var wg sync.WaitGroup
    config := disp.resq.config

    for i:=0; i<disp.maxWorkers; i++{
        worker := NewWorker(config, disp.queues, i+1)
        if worker == nil {
            return errors.New("ERROR running worker: worker is nil")
        }
        workerId := worker.String()
        workerIdChan <- workerId

        wg.Add(1)
        go func () {
            err := worker.Startup(disp, &wg, tasks)
            if err != nil {
                log.Fatalf("ERROR startup worker: %s", err)
            }
        }()
    }
    wg.Add(1)
    go disp.Dispatch(&wg)
    wg.Wait()
    return nil
}

func (disp *Dispatcher) Dispatch(wg *sync.WaitGroup){
    for {
        select {
        case workerId := <-workerIdChan:
            go func(workerId string){
              for {
                job := ReserveJob(disp.resq, disp.queues, workerId)
                if job != nil {
                  disp.jobChannel<-job
                }
              }
            }(workerId)
        case <-time.After(time.Second * time.Duration(disp.timeout)):
            log.Println("Timeout: Dispatch")
            break
        }
        break
    }
    wg.Done()
}
