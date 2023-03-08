# shift4-local-utg-helper

Shift4 Local UTG Helper is installed alongside Shift4's UTG Standalone application and acts as a CORS server and terminal ID extraction utility.

### DEVELOPMENT STATUS

- [x] Create basic executable to run the HTTP server
- [x] Turn the executable into a Windows Service using [kardianos' service framework](https://github.com/kardianos/service/)
  - [-] Optionally migrate to using [Go's SVC framework](https://github.com/golang/sys/blob/master/windows/svc/example/install.go) \[[Documentation](https://pkg.go.dev/golang.org/x/sys@v0.6.0/windows/svc/mgr#Config)\]
- [-] Research installer methods
  - [x] Raw MSI
  - [-] Paid installer (Install Shield, etc)
  - [ ] Create installer BAT
  - [ ] Self installing service using Go's SVC framework
