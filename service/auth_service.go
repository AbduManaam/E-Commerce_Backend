package service

import (
	"errors"
	"log"
	"os"
	"time"

	"backend/config"
	"backend/internal/domain"
	"backend/repository"
	"backend/utils"
	"backend/utils/email"
	"backend/utils/hash"
	"backend/utils/otp"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo   repository.UserRepository
	jwtConfig  *config.JWTConfig
	emailSvc  email.Service
	logger     *log.Logger
}

// Constructor with JWT config
func NewAuthService(
	userRepo repository.UserRepository,
	jwtConfig *config.JWTConfig,
	emailSvc email.Service,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtConfig: jwtConfig,
		emailSvc:  emailSvc,
		logger:    log.New(os.Stdout, "AuthService: ", log.LstdFlags),
	}
}


// SIGNUP with OTP
func (s *AuthService) Signup(user *domain.User) error {
	if user.Name == "" || len(user.Name) < 2 || len(user.Name) > 100 {
		return ErrInvalidInput
	}
	if !isValidEmail(user.Email) {
		return ErrInvalidInput
	}
	if len(user.Password) < 6 {
		return ErrInvalidInput
	}
	if user.Role == "" {
		user.Role = "user"
	}

	existingUser, err := s.userRepo.GetByEmail(user.Email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existingUser != nil {
		return ErrUserExists
	}

	hashed, err := hash.HashPassword(user.Password)
	if err != nil {
		return err
	}

	otpCode := otp.Generate()
	otpHash, _ := otp.HashOTP(otpCode)

	user.Password = hashed
	user.OTP = string(otpHash)
	user.OTPExpiry = otp.Expiry()
	user.IsVerified = false
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Create(user); err != nil {
		return err
	}

	if err := s.emailSvc.SendVerificationOTP(user.Email, user.Name, otpCode); err != nil {
		s.logger.Printf("Failed to send OTP to %s: %v", user.Email, err)
		return &ServiceError{Code: "EMAIL_SEND_FAILED", Msg: "Failed to send OTP email"}
	}

	return nil
}

func (s *AuthService) VerifyOTP(userEmail, otpCode string) error {
	
	if userEmail == "" || otpCode == "" {
		return ErrInvalidInput
	}
	
	user, err := s.userRepo.GetByEmail(userEmail)
	if err != nil {
		s.logger.Println("[VerifyOTP] DB error fetching user:", err)
		return err 
	}	

	if user == nil {
		s.logger.Println("[VerifyOTP] user not found:", userEmail)
		return ErrUserNotFound
	}

	if user.IsVerified {
		s.logger.Println("[VerifyOTP] already verified:", userEmail)
		return &ServiceError{
			Code: "ALREADY_VERIFIED",
			Msg:  "User already verified",
		}
	}

	if time.Now().After(user.OTPExpiry) {
		s.logger.Println("[VerifyOTP] OTP expired for:", userEmail)
		return &ServiceError{
			Code: "OTP_EXPIRED",
			Msg:  "OTP expired",
		}
	}

	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.OTP),
		[]byte(otpCode),
	); err != nil {
		s.logger.Println("[VerifyOTP] invalid OTP attempt for:", userEmail)
		return &ServiceError{
			Code: "INVALID_OTP",
			Msg:  "Invalid OTP",
		}
	}

	user.IsVerified = true
	user.OTP = ""
	user.OTPExpiry = time.Time{}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		s.logger.Println("[VerifyOTP] failed to update user:", err)
		return err
	}

	s.logger.Println("[VerifyOTP] user verified successfully:", userEmail)
	return nil
}


// LOGIN
func (s *AuthService) Login(userEmail, password string) (*domain.User, string, string, error) {
	
	if !isValidEmail(userEmail) || password == "" {
		return nil, "", "", ErrInvalidInput
	}
	
	
	user, err := s.userRepo.GetByEmail(userEmail)
	if err != nil || user == nil {
		return nil, "", "", ErrInvalidLogin
	}

	// Check if user is blocked
	if user.IsBlocked {
		return nil, "", "", ErrUserBlocked
	}
	if !user.IsVerified {
		return nil, "", "", &ServiceError{
			Code: "EMAIL_NOT_VERIFIED",
			Msg:  "Please verify your email address before logging in",
		}
	}

	if !hash.CheckPassword(password, user.Password) {
		return nil, "", "", ErrInvalidLogin
	}

	// Generate JWT tokens
	accessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Role,
		s.jwtConfig.AccessSecret,
		time.Duration(s.jwtConfig.AccessExpiry)*time.Second,
	)
	if err != nil {
		return nil, "", "", err
	}

	refreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		user.Role,
		s.jwtConfig.RefreshSecret,
		time.Duration(s.jwtConfig.RefreshExpiry)*time.Second,
	)
	if err != nil {
		return nil, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

// REFRESH TOKEN
func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := utils.ValidateRefreshToken(refreshToken, s.jwtConfig.RefreshSecret)
	if err != nil {
		return "", "", ErrInvalidLogin
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil || user == nil {
		return "", "", ErrUserNotFound
	}

	if user.IsBlocked {
		return "", "", ErrUserBlocked
	}

	newAccessToken, err := utils.GenerateAccessToken(
		user.ID,
		user.Role,
		s.jwtConfig.AccessSecret,
		time.Duration(s.jwtConfig.AccessExpiry)*time.Second,
	)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := utils.GenerateRefreshToken(
		user.ID,
		user.Role,
		s.jwtConfig.RefreshSecret,
		time.Duration(s.jwtConfig.RefreshExpiry)*time.Second,
	)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

//-----------------------------------------------------

func (s *AuthService) RefreshExpiry() int {
	return s.jwtConfig.RefreshExpiry
}


// Forgot Password with OTP
func (s *AuthService) ForgotPassword(userEmail  string) error {
	user, err := s.userRepo.GetByEmail(userEmail )
	if err != nil || user == nil {
		return nil // Don't reveal if user exists
	}

	if user.IsBlocked {
		return nil
	}

	// Generate OTP
	otpCode := otp.Generate()
    otpHash, _ := otp.HashOTP(otpCode)

	user.OTP = string(otpHash)
	user.OTPExpiry = otp.Expiry()
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// Send OTP email
	if err :=s.emailSvc.SendVerificationOTP(user.Email, user.Name, otpCode); err != nil {
		s.logger.Printf("Failed to send password reset OTP to %s: %v", user.Email, err)
		return &ServiceError{
			Code: "EMAIL_SEND_FAILED",
			Msg:  "Failed to send password reset email",
		}
	}

	return nil
}

// Reset Password with OTP
func (s *AuthService) ResetPassword(userEmail, otpCode, newPassword string) error {
	user, err := s.userRepo.GetByEmail(userEmail)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	if user.IsBlocked {
		return ErrUserBlocked
	}

	// Check OTP
	if time.Now().After(user.OTPExpiry) {
		return &ServiceError{
			Code: "OTP_EXPIRED",
			Msg:  "OTP expired",
		}
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.OTP), []byte(otpCode)); err != nil {
		return &ServiceError{
			Code: "INVALID_OTP",
			Msg:  "Invalid OTP",
		}
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return &ServiceError{
			Code: "PASSWORD_HASH_FAILED",
			Msg:  "Failed to hash password",
		}
	}

	// Update password and clear OTP
	user.Password = string(hashedPassword)
	user.OTP = ""
	user.OTPExpiry = time.Time{}
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	return nil
}

// CHANGE PASSWORD (requires current password)
func (s *AuthService) ChangePassword(
	userID uint,
	oldPassword string,
	newPassword string,
) error {

	if userID == 0 || oldPassword == "" || newPassword == "" {
		return ErrInvalidInput
	}

	if oldPassword == newPassword || len(newPassword) < 8 {
		return ErrInvalidInput
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return ErrUserNotFound
	}

	if user.IsBlocked {
		return ErrUserBlocked
	}

	if !hash.CheckPassword(oldPassword, user.Password) {
		return ErrPasswordMismatch
	}

	hashed, err := hash.HashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(userID, hashed)
}




//  basic email format validation
func isValidEmail(email string) bool {
	if len(email) < 5 || !contains(email, "@") {
		return false
	}
	return true
}
func contains(s, sub string) bool { return len(s) >= len(sub) && (s[0:len(sub)] == sub || contains(s[1:], sub)) }





// ResendVerificationEmail (if needed)
func (s *AuthService) ResendVerificationEmail(userEmail  string) error {
	user, err := s.userRepo.GetByEmail(userEmail )
	if err != nil {
		return nil // Don't reveal if user exists
	}

	if user.IsVerified {
		return &ServiceError{Code: "ALREADY_VERIFIED", Msg: "Email already verified"}
	}

	// Generate new OTP
	otpCode := otp.Generate()

otpHash, _ := otp.HashOTP(otpCode)
	user.OTP = string(otpHash)
	user.OTPExpiry = otp.Expiry()
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	// Send OTP email
	if err := s.emailSvc.SendVerificationOTP(user.Email, user.Name, otpCode); err != nil {
		s.logger.Printf("Failed to send verification OTP to %s: %v", user.Email, err)
		return &ServiceError{
			Code: "EMAIL_SEND_FAILED",
			Msg:  "Failed to send verification email",
		}
	}

	return nil
}

// // Generate OTP helper function
// func generateOTP() string {
// 	// Simple 6-digit OTP generation
// 	// In production, use crypto/rand for better security
// 	return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
// }