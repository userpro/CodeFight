package eventQ

import (
    "sync"
    "container/list"
)

const default_action_length = 100 // 操作序列最大长度

type EventQueue struct { 
    mu sync.Mutex
    e  map[string]*list.List // [token]eventlist
}

type EventEle struct {
    Token string
    Value interface{} // ActionEvent
}

func New() *EventQueue { return new(EventQueue) }

/* 需要同步加锁 */
func (eq *EventQueue)Initialize(token []string) {
    eq.mu.Lock()
    eq.e = make(map[string]*list.List)
    for _, v := range(token) {
        eq.e[v] = list.New()
    }
    eq.mu.Unlock()
}

func (eq *EventQueue)Empty() bool {
    // fmt.Println("[EventQueue-Empty] ", len(eq.e))
    if len(eq.e) == 0 { return true }
    return false
}

/* 需要同步加锁 */
func (eq *EventQueue)Remove(token string) {
    eq.mu.Lock()
    delete(eq.e, token)
    eq.mu.Unlock()
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