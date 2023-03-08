package main

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

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
	// TODO: parameterize origin
	// attach CORS headers to client response
	// clientResponse.Header().Add("Access-Control-Allow-Origin", "http://localhost:3210")
	clientResponse.Header().Add("Access-Control-Allow-Origin", "*")

	// respond to preflight request by allowing all methods and headers
	if clientRequest.Method == "OPTIONS" {
		clientResponse.Header().Add("Access-Control-Allow-Method", "*")
		clientResponse.Header().Add("Access-Control-Allow-Headers", "*")
		return
	}

	// TODO: parameterize src
	// create the request to UTG server, and handle errors
	utgRequest, err := http.NewRequest(clientRequest.Method, "https://10.0.15.19:4041/api/rest/v1/transactions/sale", clientRequest.Body)
	if handleError(err, clientResponse) {
		return
	}

	// copy headers from the client request to the utg request
	copyHeaders(clientRequest.Header, utgRequest.Header)

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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return true
	}
	return false
}

func main() {
	// send the third line of the most recent EMVTERM file
	http.HandleFunc("/terminalId", func(clientResponse http.ResponseWriter, clientRequest *http.Request) {
		terminalId, err := extractTerminalId("C:/Shift4/EMV/", ".EMVTERM")
		if handleError(err, clientResponse) {
			return
		}
		clientResponse.Write([]byte(terminalId))
	})

	http.HandleFunc("/", corsProxy)
	// TODO: parameterize listen
	// start the HTTP server and handle errors (usually invalid listening address)
	if err := http.ListenAndServe(":4040", nil); err != nil {
		// if err := http.ListenAndServe("localhost:4040", nil); err != nil {
		log.Fatal(err)
	}
}
