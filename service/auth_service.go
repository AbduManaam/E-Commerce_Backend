package service

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"backend/config"
	"backend/internal/domain"
	"backend/repository"
	"backend/utils"
	"backend/utils/email"
	"backend/utils/hash"
	"backend/utils/logging"
	"backend/utils/otp"
	validator "backend/utils/validation"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	userRepo   repository.UserRepository
	authRepo repository.AuthRepository  
	jwtConfig  *config.JWTConfig
	emailSvc  email.Service
	logger     *log.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	authRepo repository.AuthRepository,
	jwtConfig *config.JWTConfig,
	emailSvc email.Service,
) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		authRepo: authRepo,
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
	
    if !validator.IsValidPassword(user.Password){
		return ErrInvalidInput.WithContext("weak password")
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
    if err != nil {
        logging.LogWarn("login failed: repo error", nil, err, "email", userEmail)
        return nil, "", "", ErrInvalidLogin
    }

    if user == nil {
        logging.LogWarn("user not found", nil, err, "Email", userEmail)
        return nil, "", "", ErrInvalidLogin.WithContext("")
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

    // Store refresh token in database
    tokenHash := utils.HashString(refreshToken) // You need to create this helper
    expiresAt := time.Now().Add(time.Duration(s.jwtConfig.RefreshExpiry) * time.Second)
    if err := s.authRepo.SaveRefreshToken(user.ID, tokenHash, expiresAt); err != nil {
        return nil, "", "", err
    }

    return user, accessToken, refreshToken, nil
}


//Logout
func(s *AuthService)Logout(refreshToken string)error{
	if refreshToken==""{
		logging.LogWarn(
			"Logout failed: empty refresh token",
			nil,   
			nil,   
			"action", "logout",
		)
		return ErrInvalidInput
	}
	tokenHash:= utils.HashString(refreshToken)
	 return s.authRepo.DeleteRefreshToken(tokenHash)
}



// REFRESH TOKEN
func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {
    // First check if token exists in DB
    tokenHash := utils.HashString(refreshToken)
    storedToken, err := s.authRepo.GetRefreshToken(tokenHash)
    if err != nil || storedToken == nil {
        return "", "", ErrInvalidToken
    }
    
    // Check expiry
    if time.Now().After(storedToken.ExpiresAt) {
        s.authRepo.DeleteRefreshToken(tokenHash)
        return "", "", ErrInvalidToken
    }

    // Now validate the JWT token
    claims, err := utils.ValidateRefreshToken(refreshToken, s.jwtConfig.RefreshSecret)
    if err != nil {
        // Also delete invalid token from database
        s.authRepo.DeleteRefreshToken(tokenHash)
        return "", "", ErrInvalidToken
    }

    user, err := s.userRepo.GetByID(claims.UserID)
    if err != nil || user == nil {
        return "", "", ErrUserNotFound
    }

    if user.IsBlocked {
		 _ = s.authRepo.DeleteAllByUserID(user.ID)
        return "", "", ErrForbidden.WithContext("account suspended")
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

    // Delete old token and save new one
    s.authRepo.DeleteRefreshToken(tokenHash)
    newTokenHash := utils.HashString(newRefreshToken)
    expiresAt := time.Now().Add(time.Duration(s.jwtConfig.RefreshExpiry) * time.Second)
    s.authRepo.SaveRefreshToken(claims.UserID, newTokenHash, expiresAt)

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
		return nil
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
	if !validator.IsValidPassword(newPassword){
    return ErrWeakPassword

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

	if oldPassword == newPassword{
		return ErrPasswordReUse
	}
	if !validator.IsValidPassword(newPassword) {
    return ErrWeakPassword
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
// func isValidEmail(email string) bool {
// 	if len(email) < 5 || !contains(email, "@") {
// 		return false
// 	}
// 	return true
// }
// func contains(s, sub string) bool { return len(s) >= len(sub) && (s[0:len(sub)] == sub || contains(s[1:], sub)) }


 


func isValidEmail(email string) bool {
	if len(email) < 6 {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	local := parts[0]
	domain := parts[1]

	if local == "" || domain == "" {
		return false
	}

	if !strings.Contains(domain, ".") {
		return false
	}

	domainParts := strings.Split(domain, ".")
	tld := domainParts[len(domainParts)-1]

	if len(tld) < 2 {
		return false
	}

	blockedDomains := map[string]bool{
		"gmail.cm":  true,
		"gamil.com": true,
		"gmial.com": true,
		"gmial.om": true,
	}

	if blockedDomains[domain] {
		return false
	}

	return true
}



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
