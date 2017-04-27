//====================================================== file = weblite.go ====
//=  A simple HTTP server written in Go                                       =
//=   - Uses goroutines to allow parallel serving                             =
//=============================================================================
//=  Notes:                                                                   =
//=    0) Inspired by the awesome weblite.c: by Dr. Ken Christenson           =
//=       (see here: http://www.csee.usf.edu/~kchriste/tools/weblite.c)       =
//=    1) Compiles via: run `GOOS=<OS>;GOARCH=<arch val>; go build main.go`   =
//=       (see here: https://golang.org/doc/install/source#environment)       =
//=    2) Serves all the things (go takes care of the Content-Type/etc.)      =
//=    3) This should NOT be considered secure                                =
//=       (It is currently not known if go prevents directory traversals)     =
//=       (I am too lazy to do HTTPS, but could try:)                         =
//=       (HTTPS example: https://gist.github.com/denji/12b3a568f092ab951456) =
//=---------------------------------------------------------------------------=
//=  History:  ActuallyFro (2017-04-27) - v0.1.0 -- first commit              =
//=============================================================================
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
)

//----- License --------------------------------------------
var wlgLicense = `Copyright (c) 2017 Brandon Froberg

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
`

//----- Web Server Configs/Settings --------------------------------------------
func startHTTPServer(RunAmount *int, wlgPortNum int, wlgDebugging bool) *http.Server {
	wlgServerStatus := &http.Server{Addr: ":" + strconv.Itoa(wlgPortNum)}

	//A Two Mode HTTP Responder
	//==========================
	http.HandleFunc("/", func(writerForHTTPResp http.ResponseWriter, r *http.Request) {
		//---------------------------
		//Mode A: Index.html handling
		//---------------------------
		if r.URL.Path == "/" { //serve index.html if no file in the web request
			if wlgDebugging {
				log.Printf("[wlg][DEBUGGING] Client asked for \"index.html\"")
			}
			writerForHTTPResp.Header().Set("Content-Type", "text/html")
			indexWebpage, indexErr := ioutil.ReadFile("index.html")
			writerForHTTPResp.WriteHeader(http.StatusOK)
			if indexErr != nil { //no index.html
				log.Printf("[wlg]   [ERROR] 'index.html' DOES NOT EXIST!")
				helloWebpage := "<html><body>hello world!</body></html>"
				writerForHTTPResp.Header().Set("Content-Length", fmt.Sprint(len(helloWebpage)))
				io.WriteString(writerForHTTPResp, helloWebpage)
			} else { //serve index
				writerForHTTPResp.Header().Set("Content-Type", "text/html")
				writerForHTTPResp.Header().Set("Content-Length", fmt.Sprint(len(indexWebpage)))
				fmt.Fprint(writerForHTTPResp, string(indexWebpage))
			}
			*RunAmount = (*RunAmount - 1)
			//----------------------------
			//Mode B: General File Serving
			//----------------------------
		} else { //serve non-index.html file
			if wlgDebugging {
				log.Printf("[wlg][DEBUGGING] Client asked for file: %s", r.URL.Path[1:])
			}
			_, errFileToSendStat := os.Stat(r.URL.Path[1:])

			if errFileToSendStat != nil { //file does NOT exist
				if os.IsNotExist(errFileToSendStat) {
					writerForHTTPResp.Header().Set("Content-Type", "text/html")
					log.Printf("[wlg]   [ERROR] file DOES NOT EXIST! -- sending 404")
					writerForHTTPResp.WriteHeader(http.StatusNotFound)
					fmt.Fprint(writerForHTTPResp, "<html><body><font size=\"6\"/><b>FILE NOT FOUND :(</b></body></html>")

				}
			} else { // file exists
				Filename := r.URL.Path[1:]
				Openfile, _ := os.Open(Filename)
				defer Openfile.Close()
				FileHeader := make([]byte, 512)
				Openfile.Read(FileHeader)
				FileContentType := http.DetectContentType(FileHeader)

				FileStat, _ := Openfile.Stat()
				FileSize := strconv.FormatInt(FileStat.Size(), 10)

				writerForHTTPResp.Header().Set("Content-Disposition", "attachment; filename="+Filename)
				writerForHTTPResp.Header().Set("Content-Type", FileContentType)
				writerForHTTPResp.Header().Set("Content-Length", FileSize)

				Openfile.Seek(0, 0)
				io.Copy(writerForHTTPResp, Openfile) //'Copy' the file to the client
				*RunAmount = (*RunAmount - 1)
				return

			}
		}
	})

	//Run as a parallel goroutine to listen and respond to a client request:
	go func() {
		if errListen := wlgServerStatus.ListenAndServe(); errListen != nil {
			if errListen.Error() == "http: Server closed" {
				if wlgDebugging {
					log.Printf("[wlg][WARNING] The server shutdown, can't ListenAndServe()")
				}
			} else {
				log.Printf("[wlg][ERROR] ListenAndServe() error: %s", errListen)
			}
		}
	}()

	return wlgServerStatus
}

//----- Web Server Configs/Settings --------------------------------------------
func main() {
	var wlgRunAmount int
	var wlgPortNum int
	var wlgFlagRunOnce bool
	var wlgDebugging bool
	var wlgPrintLicense bool

	flag.Usage = func() {
		fmt.Printf("Weblite-Go (v0.1.0):\n====================\n")
		fmt.Print("This is a simple, but robust, webserver written in Go. It is designed to have \n")
		fmt.Print("the minimal needed functionality to serve files on the web.\n\n")
		fmt.Print("Current Flags (both - and -- can invoke flag args):\n---------------------------------------------------\n")
		flag.PrintDefaults()
	}

	flag.BoolVar(&wlgFlagRunOnce, "1", false, "Set the server to provide a single file")
	flag.BoolVar(&wlgDebugging, "debug", false, "Prints debugging messages")
	flag.BoolVar(&wlgPrintLicense, "license", false, "Prints the included license")
	flag.IntVar(&wlgRunAmount, "amount", -1, "Set the server to provide # file(s)")
	flag.IntVar(&wlgPortNum, "port", 8080, "Set the server's listening port")
	flag.Parse()

	wlgFlagSetAmount := flag.CommandLine.Lookup("amount")
	wlgFlagSetAmountInt, _ := strconv.Atoi(wlgFlagSetAmount.Value.String())

	wlgServerMode := "infinite"

	if wlgFlagSetAmountInt != -1 && len(os.Args) >= 3 {
		wlgServerMode = "finite"
	}

	if wlgDebugging {
		log.Printf("[wlg][DEBUGGING] flags: %t, %d, %s", wlgFlagRunOnce, wlgRunAmount, wlgServerMode)
	}

	if wlgFlagRunOnce {
		wlgRunAmount = 1
		wlgServerMode = "finite"
	}

	if wlgPortNum > 65535 || wlgPortNum < 1 {
		log.Printf("[wlg][ERROR] Port value (%d) is out of bounds!", wlgPortNum)
		os.Exit(-1)
	}

	if wlgPrintLicense {
		fmt.Printf("%s", wlgLicense)
		os.Exit(1)
	}
	log.Printf("[wlg] starting HTTP server")
	wlgServerStatus := startHTTPServer(&wlgRunAmount, wlgPortNum, wlgDebugging)

	if wlgServerMode == "infinite" {
		log.Printf("[wlg] serving forEVER!")

		var input string
		for {
			// text, _ := reader.ReadString('\n')
			log.Printf("[wlg]    Press 'Enter' to quit...")
			fmt.Scanf("%s", &input)

			if input == "" {
				log.Printf("[wlg]       Enter was pressed...quiting!\n")
				break
			} else {
				log.Printf("[wlg]       [WARNING] You didn't press ONLY enter...Press 'Enter' to quit...\n")
				input = ""
			}
		}
	} else { // wlgServerMode == finite
		log.Printf("[wlg] serving for %d files", wlgRunAmount)
		for {
			if wlgRunAmount <= 0 {
				break
			}
		}
		log.Printf("[wlg] stopping HTTP server")
	}
	wlgServerStatus.Shutdown(context.Background())
	log.Printf("[wlg] done. exiting")
}
