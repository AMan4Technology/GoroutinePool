package gopool

const (
    Reject Strategy = iota
    WaitReject
    Wait
    WaitEval
    Eval
)

type Strategy int
