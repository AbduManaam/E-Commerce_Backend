package email

type SMTPService struct{}

type Service interface {
	Send(to, subject, body string) error
	SendVerificationOTP(to, name, otp string) error
	SendPasswordResetOTP(to, name, otp string) error
}


func NewSMTPService() Service {
	return &SMTPService{}
}

func (s *SMTPService) Send(to, subject, body string) error {
	return sendEmail(to, subject, body)
}

func (s *SMTPService) SendVerificationOTP(to, name, otp string) error {
	return SendVerificationOTP(to, name, otp)
}

func (s *SMTPService) SendPasswordResetOTP(to, name, otp string) error {
	return SendPasswordResetOTP(to, name, otp)
}
