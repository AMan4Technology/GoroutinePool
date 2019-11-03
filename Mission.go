package gopool

import (
    "fmt"
    "time"

    "github.com/AMan4Technology/DataStructure/heap"
    "github.com/AMan4Technology/DataStructure/useful/compare"
)

func NewMission(id string, level int8, deadline time.Time, protect bool, work func()) Mission {
    return Mission{
        id:       id,
        work:     work,
        deadline: deadline,
        level:    level,
        protect:  protect,
    }
}

type Mission struct {
    id                    string
    work                  func()
    deadline, enqueueTime time.Time
    level                 int8
    protect               bool
}

func (m Mission) Compare(b heap.Value) (result int8) {
    mission := b.(Mission)
    if result = compare.Float64(float64(m.level), float64(mission.level)); result != compare.EqualTo || m.enqueueTime.Equal(mission.enqueueTime) {
        return
    }
    if m.enqueueTime.Before(mission.enqueueTime) {
        return compare.MoreThan
    }
    return compare.LessThan
}

func (m *Mission) CanEnqueue() bool {
    m.enqueueTime = time.Now()
    return m.enqueueTime.Before(m.deadline)
}

func (m Mission) Timeout() bool {
    return time.Now().After(m.deadline)
}

func (m Mission) ID() string {
    return m.id
}

func (m Mission) Eval() (err error, duration time.Duration) {
    if m.Timeout() {
        return fmt.Errorf("mission{%s} should eval at %s", m.id, m.deadline), time.Duration(0)
    }
    return m.ForceEval()
}

func (m Mission) ForceEval() (err error, duration time.Duration) {
    now := time.Now()
    defer func() { duration = time.Since(now) }()
    if m.protect {
        defer func() {
            if p := recover(); p != nil {
                err = fmt.Errorf("mission{%s} fail, error:%e", m.id, p)
            }
        }()
    }
    m.work()
    return
}
