package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"gin-quickstart/utils"
	"log"
	"net/smtp"
	"os"
)

func SendHTMLWithInlineImage(to string, subject string, htmlBody string, imgData []byte) error {
	if os.Getenv("EMAILS_ENABLED") != "true" {
		log.Println("Email sending is disabled (EMAILS_ENABLED != true). Skipping email to", to)
		return nil
	}

	from := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASS")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	if host == "" || from == "" {
		log.Println("Email service not configured. Skipping email to", to)
		return nil
	}

	boundary := "related-boundary-12345"

	message := bytes.NewBuffer(nil)
	message.WriteString(fmt.Sprintf("From: %s\r\n", from))
	message.WriteString(fmt.Sprintf("To: %s\r\n", to))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString(fmt.Sprintf("Content-Type: multipart/related; boundary=%s\r\n", boundary))
	message.WriteString("\r\n")

	message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	message.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	message.WriteString("\r\n")
	message.WriteString(htmlBody)
	message.WriteString("\r\n")

	if imgData != nil {
		message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		message.WriteString("Content-Type: image/png\r\n")
		message.WriteString("Content-Transfer-Encoding: base64\r\n")
		message.WriteString("Content-ID: <qrcode>\r\n")
		message.WriteString("Content-Disposition: inline; filename=\"qrcode.png\"\r\n")
		message.WriteString("\r\n")

		encoded := base64.StdEncoding.EncodeToString(imgData)
		message.WriteString(encoded)
		message.WriteString("\r\n")
	}

	message.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	auth := smtp.PlainAuth("", from, pass, host)
	err := smtp.SendMail(host+":"+port, auth, from, []string{to}, message.Bytes())
	if err != nil {
		log.Printf("Failed to send email to %s: %v", to, err)
		return err
	}

	return nil
}

func SendBookingConfirmation(patientEmail string, doctorName string, examinationName string) {
	subject := "Appointment Confirmation - Diagnostyka-App"
	html := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif;">
			<h2 style="color: #2c3e50;">Appointment Confirmed!</h2>
			<p>Hello,</p>
			<p>Your appointment for <strong>%s</strong> has been registered.</p>
			<p>Assigned Doctor: <strong>%s</strong></p>
			<p>Once your results are ready, you will receive another email with your QR code.</p>
			<br>
			<p>Best regards,<br>Diagnostyka-App Team</p>
		</body>
		</html>`, examinationName, doctorName)

	go SendHTMLWithInlineImage(patientEmail, subject, html, nil)
}

func SendResultsNotification(patientEmail string, resultHash string) {
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:8080"
	}

	resultsURL := fmt.Sprintf("%s/api/results/%s", appURL, resultHash)

	qrData, err := utils.GenerateQRCodePNG(resultsURL)
	if err != nil {
		log.Println("Failed to generate QR code for email:", err)
		return
	}

	subject := "Your Results are Ready! - Diagnostyka-App"
	html := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; text-align: center;">
			<h2 style="color: #27ae60;">Diagnostic Results Ready</h2>
			<p>Your results are now available for viewing.</p>
			<p>Scan the QR code below or click the link:</p>
			<div style="margin: 20px 0;">
				<img src="cid:qrcode" alt="QR Code" width="200" height="200" style="border: 1px solid #ddd; padding: 10px;">
			</div>
			<p><a href="%s" style="background-color: #27ae60; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">View Results Online</a></p>
			<br>
			<p style="color: #7f8c8d; font-size: 12px;">Link: %s</p>
		</body>
		</html>`, resultsURL, resultsURL)

	go SendHTMLWithInlineImage(patientEmail, subject, html, qrData)
}
