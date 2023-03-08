package main

import (
	"fmt"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

var elog debug.Log

type myservice struct{}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				changes <- svc.Status{State: svc.StopPending}
				elog.Info(1, fmt.Sprintf("%s service stopped", "Shift4 UTG Helper FROMLOG"))
				break loop
			}
		}
	}

	return false, 0

}

func runService(name string, isDebug bool) {
	var err error
	elog, err = eventlog.Open(name)
	if err != nil {
		return
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	err = run(name, &myservice{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}

	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}

func init() {
	elog, _ = eventlog.Open("Shift4 UTG Helper EVENTLOG")
	elog.Error(1, fmt.Sprintf("where does this log go?"))

	// runService("what is this?", true)
	go svc.Run("Shift4 UTG Helper RUN", &myservice{})
}
