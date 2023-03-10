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
	"golang.org/x/sys/windows/svc/mgr"
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
		// TODO: Fix, this doesn't shut down when the service stops
		go runService()
	} else {
		handleInput()
	}

	startServer()
}

func handleInput() {
	cmd := ""
	if len(os.Args) >= 2 {
		cmd = strings.ToLower(os.Args[1])
	}
	switch cmd {
	case "install":
		err := installSvc()
		if err != nil {
			log.Fatalf("\nERROR: %v", err)
		}
	case "uninstall":
		err := uninstallSvc()
		if err != nil {
			fmt.Printf("\nERROR: %v", err)
			// log.Fatalf("\nERROR: %v", err)
		}
	case "run", "start":
		return
	case "":
		fmt.Println("no arguments provided, starting menu")
	default:
		// Do nothing. Unrecognized args will be handled by the flag library in startServer
		return

	}

	fmt.Println("Welcome to the Shift4 UTG Helper menu.")
	fmt.Println("In the future, you can skip straight to running the service by\nproviding parameters or running with the \"start\" argument.")
	fmt.Println("For a list of available parameters and their default values,\nrun this program with the \"-help\" argument.\n")
	fmt.Println("Select an action:")
	fmt.Println("\t1. Install")
	fmt.Println("\t2. Uninstall")
	fmt.Println("\t3. Start (using default parameters)")
	fmt.Println("\t4. Exit")
	var selection int
	for {
		fmt.Print("Selection [1-4]: ")
		fmt.Scan(&selection)
		if selection >= 1 && selection <= 4 {
			break
		}
		fmt.Printf("%d is not a value between 1 and 4. Try again.\n", selection)
	}

	switch selection {
	case 1:
		err := installSvc()
		if err != nil {
			fmt.Printf("\nERROR: %v", err)
			// log.Fatalf("\nERROR: %v", err)
		}
	case 2:
		err := uninstallSvc()
		if err != nil {
			fmt.Printf("\nERROR: %v", err)
			// log.Fatalf("\nERROR: %v", err)
		}
	case 3:
		// start the server, which will be done automatically if we leave this function
	case 4:
		os.Exit(0)
	}
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

func uninstallSvc() error {
	fmt.Print("Uninstalling Shift4 UTG Helper...")

	err := os.RemoveAll("C:\\Program Files\\Shift4 Helper")

	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(svcName)
	if err != nil {
		return err
		// return fmt.Errorf("%s service is not installed", svcName)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(svcName)
	if err != nil {
		return err
		// return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}

	fmt.Println("Done")

	fmt.Println("Please reboot for changes to take effect")

	return err
}

func wait() {
	var a string
	fmt.Scan(&a)
}
