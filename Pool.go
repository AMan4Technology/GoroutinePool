package gopool

import (
    "errors"
    "fmt"
    "sync"
    "time"

    "github.com/AMan4Technology/DataStructure/heap"
)

func NewPool(name string, cap int, min, workers, max int32, freeTime time.Duration) (*Pool, error) {
    if min < 0 || workers < min || max < workers {
        return nil, errors.New("invalid arguments")
    }
    var pool = Pool{
        name:           name,
        queue:          heap.NewMax(cap),
        workerSig:      make(chan int32),
        closeSig:       make(chan struct{}),
        enqueueMission: make(chan Mission),
        dequeueMission: make(chan Mission),
        freeTime:       freeTime,
        min:            min,
        workers:        workers,
        max:            max,
    }
    pool.init()
    return &pool, nil
}

type Pool struct {
    name                           string
    queue                          heap.Max
    mu                             sync.Mutex
    closeSig                       chan struct{}
    workerSig                      chan int32
    enqueueMission, dequeueMission chan Mission
    freeTime                       time.Duration
    min, workers, max              int32
}

func (p Pool) Name() string {
    return p.name
}

func (p Pool) Workers() int32 {
    return p.workers
}

func (p Pool) Full() bool {
    return p.queue.Full()
}

func (p Pool) Empty() bool {
    return p.queue.Empty()
}

func (p Pool) UpdateWorkers(num int32) {
    if (num > 0 && p.Full() && p.workers < p.max) || (num < 0 && !p.Full() && p.workers > p.min) {
        p.workerSig <- num
    }
}

func (p *Pool) Enqueue(mission Mission) error {
    return p.EnqueueWith(mission, time.Duration(0), Wait)
}

func (p *Pool) EnqueueWith(mission Mission, timeout time.Duration, strategy Strategy) (err error) {
    p.UpdateWorkers(1)
    switch strategy {
    case Reject:
        select {
        case p.enqueueMission <- mission:
        default:
            return fmt.Errorf("reject mission{%s}", mission.id)
        }
    case WaitReject:
        after := time.NewTimer(timeout)
        defer after.Stop()
        select {
        case <-after.C:
            return fmt.Errorf("wait %s then reject mission{%s}", timeout, mission.id)
        case p.enqueueMission <- mission:
        }
    case WaitEval:
        after := time.NewTimer(timeout)
        defer after.Stop()
        select {
        case <-after.C:
            err, _ = mission.Eval()
            return
        case p.enqueueMission <- mission:
        }
    case Eval:
        select {
        case p.enqueueMission <- mission:
        default:
            err, _ = mission.Eval()
            return
        }
    default:
        p.enqueueMission <- mission
    }
    return nil
}

func (p *Pool) init() {
    go func() {
        for {
            if p.Full() {
                continue
            }
            for mission := range p.enqueueMission {
                if !mission.CanEnqueue() {
                    continue
                }
                p.mu.Lock()
                _ = p.queue.Enqueue(mission, false)
                p.mu.Unlock()
                break
            }
        }
    }()
    go func() {
        for {
            if !p.Empty() {
                p.mu.Lock()
                mission := p.queue.Dequeue().(Mission)
                p.mu.Unlock()
                p.dequeueMission <- mission
            }
        }
    }()
    p.start()
}

func (p *Pool) start() {
    p.addWorkers(p.workers)
    go func() {
        for change := range p.workerSig {
            if change > 0 {
                if canChange := p.max - p.workers; canChange <= 0 {
                    continue
                } else if change > canChange {
                    change = canChange
                }
                p.workers += change
                p.addWorkers(change)
            } else if change < 0 {
                change = -change
                if canChange := p.workers - p.min; canChange <= 0 {
                    continue
                } else if change > canChange {
                    change = canChange
                }
                p.workers -= change
                p.removeWorkers(change)
            }
        }
    }()
}

func (p Pool) addWorkers(num int32) {
    for i := 0; i < int(num); i++ {
        p.addWorker()
    }
}

func (p Pool) addWorker() {
    go func() {
        for timeout := time.NewTimer(p.freeTime); ; timeout.Reset(p.freeTime) {
            select {
            case mission := <-p.dequeueMission:
                if err, duration := mission.Eval(); err == nil {
                    fmt.Println(duration)
                }
            case <-timeout.C:
                p.UpdateWorkers(-1)
            case <-p.closeSig:
                timeout.Stop()
                return
            }
        }
    }()
}

func (p Pool) removeWorkers(num int32) {
    for i := 0; i < int(num); i++ {
        p.removeWorker()
    }
}

func (p Pool) removeWorker() {
    p.closeSig <- struct{}{}
}
