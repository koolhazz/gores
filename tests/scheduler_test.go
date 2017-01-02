package tests

import (
    "fmt"
    "strconv"
    "testing"
    "github.com/wang502/gores/gores"
)

var (
    basic_sche = gores.NewScheduler()
    resq = gores.NewResQ()
    args = map[string]interface{}{"id": 1}
    item = gores.TestItem{
             Name: "TestItem",
             Queue: "TestScheduler",
             Args: args,
             Enqueue_timestamp: resq.CurrentTime(),
             Retry: true,
             Retry_every: 10,
           }
)

func TestNewScheduler(t *testing.T){
    sche := gores.NewScheduler()
    if sche == nil {
        t.Errorf("ERROR initialize Scheduler")
    }
}

func TestHandleDelayedItems(t *testing.T){
    // enqueue item to delayed queue
    err := resq.Enqueue_at(1483079527, item)
    if err != nil {
        t.Errorf("ERROR Enqueue item at timestamp %d", 1483079527)
    }
    basic_sche.Run()

    delayed_queue_size := resq.SizeOfQueue(fmt.Sprintf(gores.DEPLAYED_QUEUE_PREFIX, strconv.FormatInt(1483079527, 10)))
    if delayed_queue_size != 0 {
        t.Errorf("Scheduler worker did not handle delayed items")
    }

    queue_size := resq.Size(item.Queue)
    if queue_size != 1 {
        t.Errorf("Scheduler worker did not enqueue delayed item to resq:queue:%s", item.Queue)
    }

    resq.Pop(item.Queue)
}
