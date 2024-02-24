package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

// App contains configuration parameters for the web server
type App struct {
	Config      Configuration
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Get HTTP headers
	var headerAsString string
	for name, values := range r.Header {
		for _, value := range values {
			headerAsString += name + ": " + value + "\n"
		}
	}

	// get Body
	var requestBody string
	if r.Method != http.MethodGet {
		// get post Body
		b, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatalln(err)
		}
		requestBody = string(b)
	}

	a.ErrorLogger.Println("Incoming URL: " + r.RequestURI + "\nClient IP: " +
		r.RemoteAddr + "\n==Header==:\n" + headerAsString + "\n==Body==:\n" + requestBody)

	// check if configured token has been requested
	token, err := Find(r.RequestURI)
	if err != nil {
		a.ErrorLogger.Println("Could not find token due to error: ", err)
		// don't abort here, just behave as normal to prevent token enumeration
	} else {
		a.ErrorLogger.Printf("Found token: %d\n", token.ID)
		msg := "Token Title: " + token.Title +
			"\r\nToken Comment: " + token.Comment +
			"\r\nRequested URL: " + r.RequestURI +
			"\r\nClient IP: " + r.RemoteAddr +
			"\r\n\r\n==Header==\r\n" + headerAsString +
			"\r\n==Body==\r\n" + requestBody
		go Alert(a.Config.SMTPServer, a.Config.SMTPPort, a.Config.SMTPUser, a.Config.SMTPPassword, token.NotifyReceiver, msg)
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
