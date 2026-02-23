package email

import (
	"fmt"
	"net/smtp"

	"backend/config"
	"backend/utils/logging"
)

var smtpCfg config.SMTPConfig

func Init(cfg config.SMTPConfig) {
	smtpCfg = cfg
	logging.LogInfo("SMTP initialized", "host", smtpCfg.Host)
}

func sendEmail(to, subject, body string) error {
	from := smtpCfg.Username
	fromHeader := smtpCfg.From
	if fromHeader == "" {
		fromHeader = from
	}

	logging.LogDebug("sending email", "to", to, "host", smtpCfg.Host)

	msg := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-Version: 1.0\r\n"+
			"Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n"+
			"%s",
		fromHeader, to, subject, body,
	)

	addr := fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port)

	auth := smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host)

	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
	if err != nil {
		logging.LogError("failed to send email", "to", to, "error", err)
		return err
	}

	logging.LogInfo("email sent successfully", "to", to)
	return nil
}

// SendOTP sends an OTP email (generic function)
func SendOTP(to string, otp string) error {
	body := fmt.Sprintf(
		"Hello!\n\nYour OTP code is: %s\nIt will expire in 5 minutes.\n\nThanks,\nBackend Team",
		otp,
	)
	return sendEmail(to, "Your OTP Code", body)
}

// SendVerificationOTP sends verification OTP
func SendVerificationOTP(to, name, otp string) error {
	body := fmt.Sprintf(
		"Hello %s,\n\n"+
			"Your account verification OTP is: %s\n"+
			"Enter this code to verify your email address.\n\n"+
			"This OTP will expire in 5 minutes.\n\n"+
			"Thanks,\nBackend Team",
		name, otp,
	)
	return sendEmail(to, "Verify Your Email - OTP", body)
}

// SendPasswordResetOTP sends password reset OTP
func SendPasswordResetOTP(to, name, otp string) error {
	body := fmt.Sprintf(
		"Hello %s,\n\n"+
			"Your password reset OTP is: %s\n"+
			"Enter this code to reset your password.\n\n"+
			"This OTP will expire in 15 minutes.\n\n"+
			"If you didn't request this, please ignore this email.\n\n"+
			"Thanks,\nBackend Team",
		name, otp,
	)
	return sendEmail(to, "Password Reset - OTP", body)
}

// SendVerificationEmail sends email verification link (for backward compatibility)
func SendVerificationEmail(to, name, token string) error {
	verificationLink := fmt.Sprintf("http://localhost:8080/auth/verify-email?token=%s", token)

	body := fmt.Sprintf(
		"Hello %s,\n\n"+
			"Please verify your email address by clicking the link below:\n\n"+
			"%s\n\n"+
			"This link will expire in 24 hours.\n\n"+
			"Thanks,\nBackend Team",
		name, verificationLink,
	)
	return sendEmail(to, "Verify Your Email Address", body)
}

// SendPasswordResetEmail sends password reset link (for backward compatibility)
func SendPasswordResetEmail(to, name, token string) error {
	resetLink := fmt.Sprintf("http://localhost:8080/auth/reset-password?token=%s", token)

	body := fmt.Sprintf(
		"Hello %s,\n\n"+
			"Reset your password using the link below (valid for 1 hour):\n\n"+
			"%s\n\n"+
			"If you didn't request this, ignore this email.\n\n"+
			"Thanks,\nBackend Team",
		name, resetLink,
	)
	return sendEmail(to, "Password Reset", body)
}

// Send generic email
func Send(to, subject, body string) error {
	return sendEmail(to, subject, body)
}
