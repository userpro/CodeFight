package eventQ

import (
    "container/list"
)

const default_action_length = 100

type EventQueue struct { 
    e map[string]*list.List // [token]eventlist
}

type EventEle struct {
    Token string
    Value interface{} // ActionEvent
}

func New() *EventQueue { return new(EventQueue) }

func (eq *EventQueue)Initialize(token []string) {
    eq.e = make(map[string]*list.List)
    for _, v := range(token) {
        eq.e[v] = list.New()
    }
}

func (eq *EventQueue)Empty() bool {
    // fmt.Println("[EventQueue-Empty] ", len(eq.e))
    if len(eq.e) == 0 { return true }
    return false
}

func (eq *EventQueue)Remove(token string) {
    delete(eq.e, token)
}

func (eq *EventQueue)Push(token string, value interface{}) int {
    t, ok := eq.e[token]
    if !ok { return 0 }
    if t.Len() > default_action_length { return 0 }
    t.PushBack(value)
    return t.Len()
}

func (eq *EventQueue)Get() []EventEle {
    var t []EventEle
    for k, v := range eq.e {
        if v.Len() > 0 {
            t = append(t, EventEle{k, v.Front().Value})
            v.Remove(v.Front())
        }
    }
    return t;
}

func (eq *EventQueue)Clear() {
    for _, v := range eq.e {
        v.Init()
    }
}