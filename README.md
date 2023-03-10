# shift4-local-utg-helper

Shift4 Local UTG Helper is installed alongside Shift4's UTG Standalone application and acts as a CORS server and terminal ID extraction utility.

# build

`rsrc -manifest utg-helper.exe.manifest -ico icon.ico -o utg-helper.syso`
`GOOS=windows GOARCH=amd64 go build -o bin/utg-helper.exe`

### DEVELOPMENT STATUS

- [x] Create basic executable to run the HTTP server
- [x] Turn the executable into a Windows Service using [kardianos' service framework](https://github.com/kardianos/service/)
  - [x] Optionally migrate to using [Go's SVC framework](https://github.com/golang/sys/blob/master/windows/svc/example/install.go) \[[Documentation](https://pkg.go.dev/golang.org/x/sys@v0.6.0/windows/svc/mgr#Config)\]
- [x] Research installer methods
  - [x] Raw MSI
  - [x] Paid installer (Install Shield, etc)
  - [x] Create installer BAT
  - [x] Self installing service using Go's SVC framework
- [x] Accept flags
- [ ] Gracefully shut down http server
- [x] Create CLI Installer
- [-] Handle parameters and menu
