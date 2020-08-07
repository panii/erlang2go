package erlang

import(
    "runtime"
)

type LoopSig bool
type Atom int
type Tuple []interface{}
type Process chan Tuple
type Patten map[Atom]func(...interface{})
type PattenLoop map[Atom]func(...interface{}) LoopSig

var defaultWorks = runtime.NumCPU()

const LoopEnd LoopSig = false
const LoopContinue LoopSig = true

func Spawn(Fun func(Process)) Process {
    var Pid = make(Process)
    go Fun(Pid)

    return Pid
}

func Receive(Self Process, Patten Patten) {
    var Tuple = <-Self
    // Patten[Tuple[0].(Atom)](Tuple[1:]...)
    Patten[Tuple[0].(Atom)](Tuple...)
}

func LoopReceive(Self Process, Patten PattenLoop) {
    for Tuple := range Self {
        // if Patten[Tuple[0].(Atom)](Tuple[1:]...) == LoopEnd {
        if Patten[Tuple[0].(Atom)](Tuple...) == LoopEnd {
            break
        }
    }
}

func GetProcess(args []interface{}) Process {
    return args[len(args) - 1].(Process)
}

func Send(Process Process, args ...interface{}) {
    if len(args) > 0 && args[0] != nil {
        Tuple := make(Tuple, 0, len(args))
        for _, V := range args {
            Tuple = append(Tuple, V)
        }
		Process <- Tuple
	}
}

func Self() Process {
    var Pid = make(Process)
    return Pid
}

func (Self Process) Receive(Patten Patten) {
    Receive(Self, Patten)
}

func (Self Process) Send(Process Process, args ...interface{}) {
    args = append(args, Self)
    Send(Process, args...)
}

func (Self Process) LoopReceive(Patten PattenLoop) {
    LoopReceive(Self, Patten)
}