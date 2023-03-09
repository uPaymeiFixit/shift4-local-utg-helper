package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
)

const svcName = "Shift4 UTG Helper"
const eventid uint32 = 44227

var elog *eventlog.Log

type winservice struct{}

func (m *winservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
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
				elog.Info(eventid, fmt.Sprintf("%s service stopped", svcName))
				break loop
			default:
				elog.Error(eventid, fmt.Sprintf("Unexpected control request #%d", c))
			}
		}
	}

	return
}

func runService() {
	elog.Info(eventid, fmt.Sprintf("Starting %s service.", svcName))

	if err := svc.Run(svcName, &winservice{}); err != nil {
		elog.Error(eventid, fmt.Sprintf("%s service failed: %v", svcName, err))
		return
	}

	elog.Info(eventid, fmt.Sprintf("%s service stopped.", svcName))
}

func main() {
	elog, _ = eventlog.Open(svcName)
	defer elog.Close()

	inService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if we are running in service: %v", err)
	}
	if inService {
		go runService()
	}

	startServer()
}
