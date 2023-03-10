package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

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
func installSvc() error {

	fmt.Print("Installing Shift4 UTG Helper...")

	exepath, err := exePath()
	if err != nil {
		return err
	}
	installpath := "C:\\Program Files\\Shift4 Helper\\utg-helper.exe"
	err = copyFile(exepath, installpath)
	if err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(svcName)
	if err == nil {
		s.Close()
		return fmt.Errorf("%s service already exists", svcName)
	}
	s, err = m.CreateService(svcName, installpath, mgr.Config{DisplayName: svcName, Dependencies: []string{"frmUtg2Service"}}, "is", "auto-started")
	if err != nil {
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(svcName, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}

	fmt.Println("Done")
	fmt.Printf("\nTo configure the service, open services.msc and navigate to \"%s\".  ", svcName)

	// Keep the window open so the user can read
	wait()
	os.Exit(0)
	return nil
}

// return the path to this executable
func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

func copyFile(src string, dest string) (err error) {
	path := "C:\\Program Files\\Shift4 Helper"
	os.Mkdir(path, os.ModePerm)

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}
