# Write goroutine in erlang style

```go
package erlang_style_safemap

import (
    "github.com/panii/erlang2go/erlang"
)

type SafeMap interface {
    Insert(string, interface{})
    Delete(string)
    Find(string) (interface{}, bool)
    Len() int
    Update(string, UpdateFunc)
    Close() map[string]interface{}
}

type UpdateFunc func(interface{}, bool) interface{}

type safeMapProcess erlang.Process

const (
    remove erlang.Atom = iota
    end
    find
    insert
    length
    update
    result
)


// API
func New() SafeMap {
    var Pid = erlang.Spawn(func(Self erlang.Process) {
        run(Self)
    })
    
    return safeMapProcess(Pid)
}

func (Pid safeMapProcess) Insert(Key string, Value interface{}) {
    erlang.Send(erlang.Process(Pid), insert, Key, Value)
}

func (Pid safeMapProcess) Delete(Key string) {
    erlang.Send(erlang.Process(Pid), remove, Key)
}

func (Pid safeMapProcess) Find(Key string) (value interface{}, found bool) {
    var Self = erlang.Self()
    
    Self.Send(erlang.Process(Pid), find, Key)
    
    Self.Receive(erlang.Patten{
       result: func(args ...interface{}) {
           value = args[1]
           found = args[2].(bool)
       },
    })
    
    return
}

func (Pid safeMapProcess) Len() (len int) {
    var Self = erlang.Self()
    
    Self.Send(erlang.Process(Pid), length)
    
    Self.Receive(erlang.Patten{
       result: func(args ...interface{}) {
           len = args[1].(int)
       },
    })
    
    return
}

func (Pid safeMapProcess) Update(Key string, UpdateFunc UpdateFunc) {
    erlang.Send(erlang.Process(Pid), update, Key, UpdateFunc)
}

func (Pid safeMapProcess) Close() (Store map[string]interface{}) {
    var Self = erlang.Self()
    
    Self.Send(erlang.Process(Pid), end)
    
    Self.Receive(erlang.Patten{
       result: func(args ...interface{}) {
           Store = args[1].(map[string]interface{})
       },
    })
    
    return
}

// internal
func run(Self erlang.Process) {
    var Store = make(map[string]interface{})
    
    Self.LoopReceive(erlang.PattenLoop{
        insert: func(args ...interface{}) erlang.LoopSig {
            var Key = args[1].(string)
            var Value = args[2]
            Store[Key] = Value
            
            return erlang.LoopContinue
        },
        remove: func(args ...interface{}) erlang.LoopSig {
            var Key = args[1].(string)
            delete(Store, Key)
            
            return erlang.LoopContinue
        },
        find: func(args ...interface{}) erlang.LoopSig {
            var Key = args[1].(string)
            var Parent = args[2].(erlang.Process)
            Value, Found := Store[Key]
            
            erlang.Send(Parent, result, Value, Found)
            
            return erlang.LoopContinue
        },
        length: func(args ...interface{}) erlang.LoopSig {
            var Parent = args[1].(erlang.Process)
            
            erlang.Send(Parent, result, len(Store))
            
            return erlang.LoopContinue
        },
        update: func(args ...interface{}) erlang.LoopSig {
            var Key = args[1].(string)
            var Updater = args[2].(UpdateFunc)
            Value, Found := Store[Key]
            Store[Key] = Updater(Value, Found)
            
            return erlang.LoopContinue
        },
        end: func(args ...interface{}) erlang.LoopSig {
            close(Self)
            var Parent = args[1].(erlang.Process)
            
            erlang.Send(Parent, result, Store)
            
            return erlang.LoopEnd
        },
    })
}
```

```go
package golang_safemap

type SafeMap interface {
    Insert(string, interface{})
    Delete(string)
    Find(string) (interface{}, bool)
    Len() int
    Update(string, UpdateFunc)
    Close() map[string]interface{}
}

type UpdateFunc func(interface{}, bool) interface{}

type safeMap chan commandData

type commandData struct {
    action commandAction
    key string
    value interface{}
    result chan<- interface{}
    data chan<- map[string]interface{}
    updater UpdateFunc
}

type commandAction int

const (
    remove commandAction = iota
    end
    find
    insert
    length
    update
)

type findResult struct {
    value interface{}
    found bool
}

// API
func New() SafeMap {
    sm := make(safeMap)
    go sm.run()
    
    return sm
}

func (sm safeMap) Insert(key string, value interface{}) {
    sm <- commandData{action: insert, key: key, value: value}
}

func (sm safeMap) Delete(key string) {
    sm <- commandData{action: remove, key: key}
}

func (sm safeMap) Find(key string) (value interface{}, found bool) {
    reply := make(chan interface{})
    sm <- commandData{action: find, key: key, result: reply}
    result := (<-reply).(findResult)
    
    return result.value, result.found
}

func (sm safeMap) Len() int {
    reply := make(chan interface{})
    sm <- commandData{action: length, result: reply}
    
    return (<-reply).(int)
}

func (sm safeMap) Update(key string, updater UpdateFunc) {
    sm <- commandData{action: update, key: key, updater: updater}
}

func (sm safeMap) Close() map[string]interface{} {
    reply := make(chan map[string]interface{})
    sm <- commandData{action: end, data: reply}
    
    return <-reply
}

func (sm safeMap) run() {
    store := make(map[string]interface{})
    for command := range sm {
        switch command.action {
        case insert:
            store[command.key] = command.value
        case remove:
            delete(store, command.key)
        case find:
            value, found := store[command.key]
            command.result <- findResult{value, found}
        case length:
            command.result <- len(store)
        case update:
            value, found := store[command.key]
            store[command.key] = command.updater(value, found)
        case end:
            close(sm)
            command.data <- store
            break
        }
    }
}
```
