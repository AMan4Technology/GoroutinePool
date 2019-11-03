package gopool

import (
    "fmt"
    "testing"
    "time"
)

func TestPool(t *testing.T) {
    pool, _ := NewPool("test", 10, 1, 1, 1, 5*time.Second)
    _ = pool.Enqueue(NewMission("1", 1, time.Now().Add(1*time.Second), true, func() {
        fmt.Println("hello world 1")
        time.Sleep(1 * time.Second)
    }))
    _ = pool.Enqueue(NewMission("2", 1, time.Now().Add(5*time.Second), true, func() {
        fmt.Println("hello world 2")
        time.Sleep(1 * time.Second)
    }))
    _ = pool.Enqueue(NewMission("3", 1, time.Now().Add(5*time.Second), true, func() {
        fmt.Println("hello world 3")
        time.Sleep(1 * time.Second)
    }))
    _ = pool.Enqueue(NewMission("4", 2, time.Now().Add(5*time.Second), true, func() {
        fmt.Println("hello world 4")
    }))
    _ = pool.Enqueue(NewMission("5", 2, time.Now().Add(5*time.Second), true, func() {
        fmt.Println("hello world 5")
    }))
    _ = pool.Enqueue(NewMission("9", 9, time.Now().Add(5*time.Second), true, func() {
        fmt.Println("hello world 9")
    }))
    time.Sleep(5 * time.Second)
}
