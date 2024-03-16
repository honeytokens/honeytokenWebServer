package main

import (
	"net/smtp"
	"strconv"
)

// Alert sends an alert via email
func Alert(appConfig *App, receiver, message string) {
	user := appConfig.Config.SMTPUser
	password := appConfig.Config.SMTPPassword
	smtpHost := appConfig.Config.SMTPServer
	smtpPort := appConfig.Config.SMTPPort

	// Receiver email address.
	to := []string{
		receiver,
	}

	mailHeader := "From: " + user + "\r\n" +
		"To: " + receiver + "\r\n" +
		"Subject: Honeytoken triggered!\r\n" +
		"\r\n"

	// Authentication.
	auth := smtp.PlainAuth("", user, password, smtpHost)

	// Sending email.
	err := smtp.SendMail(smtpHost+":"+strconv.Itoa(smtpPort), auth, user, to, []byte(mailHeader+message))
	if err != nil {
		appConfig.ErrorLogger.Println(err)
		return
	}
	appConfig.DebugLogger.Println("Email sent successfully to " + receiver + "!")
}
