package main

import (
	"fmt"
	"net/smtp"
	"strconv"
)

func Alert(smtpHost string, smtpPort int, user, password, receiver, message string) {

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
		fmt.Println(err)
		return
	}
	fmt.Println("Email Sent Successfully!")
}
