package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

var originURL string = "*"
var utgBaseURL string = "https://localhost:4041"
var listenAddr string = "localhost:4040"
var utgInstallDir string = "C:\\Shift4\\"

func extractTerminalId(dir string, ext string) (string, error) {
	// get an array of files in our specified directory, and handle errors
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	// filter out files that don't have the specified extension
	var emvFiles []fs.FileInfo
	for i := 0; i < len(files); i++ {
		if strings.HasSuffix(strings.ToUpper(files[i].Name()), ext) {
			emvFiles = append(emvFiles, files[i])
		}
	}

	// sort our files by most recent
	sort.Slice(emvFiles, func(i, j int) bool {
		return emvFiles[i].ModTime().Before(emvFiles[j].ModTime())
	})

	// if we don't have any files, return an error
	if len(emvFiles) == 0 {
		return "", fmt.Errorf("No files with extension %v in %v", ext, dir)
	}

	// read the contents of the file, and handle errors
	fileBytes, err := ioutil.ReadFile(dir + emvFiles[0].Name())
	if err != nil {
		return "", err
	}

	// get and return the third line of the file
	fileLines := strings.Split(string(fileBytes), "\n")
	if len(fileLines) < 3 {
		return "", fmt.Errorf("File %s has less than 3 lines", emvFiles[0].Name())
	}

	return fileLines[2], err
}

func corsProxy(clientResponse http.ResponseWriter, clientRequest *http.Request) {
	// attach CORS headers to client response
	clientResponse.Header().Add("Access-Control-Allow-Origin", originURL)

	// respond to preflight request by allowing all methods and headers
	if clientRequest.Method == "OPTIONS" {
		clientResponse.Header().Add("Access-Control-Allow-Method", "*")
		clientResponse.Header().Add("Access-Control-Allow-Headers", "*")
		return
	}

	// create the request to UTG server, and handle errors
	utgRequest, err := http.NewRequest(clientRequest.Method, utgBaseURL+clientRequest.URL.String(), clientRequest.Body)
	if handleError(err, clientResponse) {
		return
	}

	// copy headers from the client request to the utg request
	copyHeaders(utgRequest.Header, clientRequest.Header)

	// make the request to the UTG server and handle errors
	utgResponse, err := http.DefaultClient.Do(utgRequest)
	if handleError(err, clientResponse) {
		return
	}

	// copy headers from the UTG response to our client response
	copyHeaders(clientResponse.Header(), utgResponse.Header)

	// send the status code
	clientResponse.WriteHeader(utgResponse.StatusCode)

	// copy the UTG response to the client response
	_, err = io.Copy(clientResponse, utgResponse.Body)
	if handleError(err, clientResponse) {
		return
	}
}

func copyHeaders(dst http.Header, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func handleError(err error, w http.ResponseWriter) bool {
	if err != nil {
		log.Printf("ERROR: %v", err)
		elog.Info(eventid, fmt.Sprintf("ERROR: %v", err))
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return true
	}
	return false
}

// func startServer(listenAddr string, utgBaseUrl string, originUrl string) {
func startServer() {
	getFlags()

	// TODO: we should not do this in production
	http.DefaultClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// send the third line of the most recent EMVTERM file
	log.Printf("Serving UTG's currently configured terminal ID on http://%s/terminalId", listenAddr)
	http.HandleFunc("/terminalId", func(clientResponse http.ResponseWriter, clientRequest *http.Request) {
		terminalId, err := extractTerminalId(utgInstallDir+"EMV\\", ".EMVTERM")
		if handleError(err, clientResponse) {
			return
		}
		clientResponse.Header().Add("Access-Control-Allow-Origin", originURL)
		clientResponse.Write([]byte(terminalId))
	})

	http.HandleFunc("/", corsProxy)
	// start the HTTP server and handle errors (usually invalid listening address)
	log.Printf("Forwarding calls originating from %s through http://%s to %s", originURL, listenAddr, utgBaseURL)
	if err := http.ListenAndServe(listenAddr, nil); err != nil {
		elog.Error(eventid, fmt.Sprintf("%s encountered an unrecoverable error: %v", svcName, err))
		log.Fatal(err)
	}
}

func getFlags() {
	flag.StringVar(&listenAddr, "listenAddr", listenAddr, "host:port this server should listen on. e.g :4040 or localhost:8080")
	flag.StringVar(&utgBaseURL, "utgBaseURL", utgBaseURL, "The base URL the Shift4 UTG server is running on. e.g. https://localhost:4041")
	flag.StringVar(&originURL, "originURL", originURL, "URL your browser will be calling from to allow CORS. e.g. https://mywebsite.com or *")
	flag.StringVar(&utgInstallDir, "utgInstallDir", utgInstallDir, "Directory Shift4's UTG software is installed in. Usually C:/Shift4/")
	flag.Parse()
}

// Install this service with:
// SC CREATE "Shift4 UTG Helper" binpath="C:\Shift4 Helper\utg-helper.exe -listenAddr=localhost:4040 -utgBaseURL=https://localhost:4041 -originURL=*" depend="frmUtg2Service"
// SC CREATE "Shift4 UTG Helper" binpath="C:\Shift4 Helper\utg-helper.exe" depend="frmUtg2Service"
