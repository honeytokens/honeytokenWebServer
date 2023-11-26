package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"os"
)

// App contains configuration parameters for the web server
type App struct {
	Config      Configuration
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestDump, _ := httputil.DumpRequest(r, true)
	defer r.Body.Close()

	a.ErrorLogger.Println("Incoming URL: " + r.RequestURI + " Client IP: " + r.RemoteAddr)
	a.DebugLogger.Println("Serving incoming request:\n", string(requestDump))

	token, err := Find(r.RequestURI)
	if err != nil {
		a.ErrorLogger.Println("Could not find token due to error: ", err)
		// don't abort here, just behave as normal to prevent token enumeration
	} else {
		a.ErrorLogger.Printf("Found token: %d\n", token.ID)
		msg := "Incoming URL: " + r.RequestURI + "\r\n" +
			"Client IP: " + r.RemoteAddr + "\r\n"
		go Alert(a.Config.SmtpServer, a.Config.SmtpPort, a.Config.SmtpUser, a.Config.SmtpPassword, token.NotifyReceiver, msg)
	}

	response, err := readResponseFromFile(a.Config.ResponseFile)
	if err != nil {
		a.ErrorLogger.Println("Could not read "+a.Config.ResponseFile+" file due to error:", err)
	}
	w.Header().Set("Content-Type", a.Config.ResponseContentType)
	w.WriteHeader(a.Config.ResponseCode)
	w.Write(response)
}

func readResponseFromFile(filename string) ([]byte, error) {
	response, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return response, nil
}
