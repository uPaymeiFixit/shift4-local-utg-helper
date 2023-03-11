package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
)

const svcName = "Shift4 UTG Helper"
const eventid uint32 = 44227

var installDir = "C:\\Program Files\\Shift4 Helper"
var serviceDependencies = []string{"frmUtg2Service"}
var server *http.Server
var elog *eventlog.Log
var inService bool

type winservice struct{}

func main() {
	elog, _ = eventlog.Open(svcName)
	defer elog.Close()

	// Determine whether this process is run as a Windows Service
	var err error
	inService, err = svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if we are running in service: %v", err)
	}

	if inService {
		// Begin the Windows Service handler
		if err := svc.Run(svcName, &winservice{}); err != nil {
			elog.Error(eventid, fmt.Sprintf("%s service failed: %v", svcName, err))
			return
		}
	} else {
		// Start the CLI menu
		cmd := ""
		if len(os.Args) >= 2 {
			cmd = strings.ToLower(os.Args[1])
		}

		if cmd == "" {
			fmt.Println("Welcome to the Shift4 UTG Helper menu.\n")
			fmt.Println("In the future, you can skip the menu by including \na \"start\", \"install\", or \"uninstall\" argument.")
			fmt.Println("For a list of additional parameters and their default values,\nrun this program with the \"-help\" argument.")
		}

		handleInput(cmd)
	}
}

// This function will be called by the Windows Service handler
func (m *winservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	elog.Info(eventid, fmt.Sprintf("Starting %s service", svcName))
	server = utgHelper()
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
				// Attempt to gracefully shut down the HTTP server
				server.Shutdown(context.Background())
				elog.Info(eventid, fmt.Sprintf("%s service stopped", svcName))
				break loop
			default:
				elog.Error(eventid, fmt.Sprintf("Unexpected control request #%d", c))
			}
		}
	}

	return
}

// CLI menu loop
func handleInput(cmd string) {
	// Handle both numeric user inputs and executable arguments
	switch cmd {
	case "1", "install":
		err := installSvc()
		if err != nil {
			fmt.Printf("Error\n  %v\n", err)
		}
	case "2", "uninstall":
		err := uninstallSvc()
		if err != nil {
			fmt.Printf("Error\n  %v\n", err)
		}
	case "":
		// This case will only be valid when the user does not provide any arguments during launch
		break
	case "4":
		os.Exit(0)
	default:
		// During our CLI control loop, this can only be 3
		// At startup we will start the server if the user provides arguments
		utgHelper()
	}

	// Print the menu
	fmt.Println("\nSelect an action:")
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

	fmt.Println()
	handleInput(strconv.Itoa(selection))
}

// Copy this executable to Program Files and add as a service
func installSvc() error {
	fmt.Print("Installing Shift4 UTG Helper...")

	// Get the path of the current executable to be copied
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	// Copy this file to installExePath
	installExePath := installDir + "\\utg-helper.exe"
	err = copyFile(exePath, installExePath)
	if err != nil {
		return err
	}

	// Begin the process of adding the Windows Service
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
	s, err = m.CreateService(svcName, installExePath, mgr.Config{DisplayName: svcName, Dependencies: serviceDependencies}, "is", "auto-started")
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

	return nil
}

// A simple function that copies files from src to dst and creates directories along the way
func copyFile(src string, dst string) (err error) {
	os.MkdirAll(installDir, os.ModePerm)

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dst)
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

// Delete the folder in Program Files and remove the Windows Service
func uninstallSvc() error {
	fmt.Print("Uninstalling Shift4 UTG Helper...")

	// Delete the folder and its contents
	err := os.RemoveAll(installDir)

	// Begin the process of removing the Windows Service
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(svcName)
	if err != nil {
		return err
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(svcName)
	if err != nil {
		return err
	}

	fmt.Println("Done")
	fmt.Println("\nPlease reboot for changes to take effect")

	return err
}
