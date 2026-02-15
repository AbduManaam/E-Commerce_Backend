package handler

import (
	"backend/handler/dto"
	"backend/internal/domain"
	"backend/service"
	"backend/utils/logging"
	validator "backend/utils/validation"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}


// POST /auth/signup
func (h *AuthHandler) Signup(c *fiber.Ctx) error {
	var req struct {
		Name     string `json:"name" validate:"required,min=2"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
	}
	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("signup body parse failed", c, err)
		return HandleError(c, service.ErrInvalidInput.WithContext("parsing signup request"))
	}

	if err:= validator.Validate.Struct(req);err!=nil{
		return c.Status(400).JSON(fiber.Map{
			"errors":validator.FormatErrors(err),
		})
	}

	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	if err := h.authSvc.Signup(user); err != nil {
		logging.LogWarn("signup failed", c, err, "email", req.Email)
		return HandleError(c, err)
	}

	logging.LogInfo("signup successful", c, "email", req.Email)
	return c.JSON(fiber.Map{"message": "OTP has been sent"})
}

// POST /auth/login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("login body parse failed", c, err)
		return HandleError(c, service.ErrInvalidInput)
	}
fmt.Println("Parsed Email:", req.Email)
	fmt.Println("Parsed Password:", req.Password)

	if err := validator.Validate.Struct(req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"errors": validator.FormatErrors(err),
		})
	}

	user, accessToken, refreshToken, err := h.authSvc.Login(req.Email, req.Password)
	if err != nil {
		logging.LogWarn("login failed", c, err, "email", req.Email)
		return HandleError(c, err)
	}

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   false,              
		SameSite: fiber.CookieSameSiteStrictMode,
		Expires:  time.Now().Add(time.Duration(h.authSvc.RefreshExpiry()) * time.Second),
		Path:     "/",
	})

	logging.LogInfo("login successful", c, "userID", user.ID)

	return c.JSON(fiber.Map{
		"message":"login successful",
		"access_token":accessToken,
		"user": dto.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
		},
	})
}



// POST /auth/refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	fmt.Println("referesh",refreshToken)
	if refreshToken == "" {
		logging.LogWarn("missing refresh token cookie", c, fiber.ErrUnauthorized)
		return HandleError(c, service.ErrUnauthorized)
	}

	accessToken, newRefreshToken, err := h.authSvc.RefreshToken(refreshToken)
	if err != nil {
		logging.LogWarn("refresh token failed", c, err)
		return HandleError(c, err)
	}

	//  ROTATE refresh token
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		HTTPOnly: true,
		Secure:   true,
		SameSite: fiber.CookieSameSiteStrictMode,
		Expires:  time.Now().Add(time.Duration(h.authSvc.RefreshExpiry()) * time.Second),
		Path:     "/auth/refresh",
	})

	return c.JSON(fiber.Map{
		"access_token": accessToken,
	})
}


//ChangePassword
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		logging.LogWarn("change password failed: no userID in context", c, service.ErrForbidden)
		return HandleError(c, service.ErrForbidden)
	}

	var req struct {
		CurrentPassword string `json:"current_password" validate:"required"`
		NewPassword     string `json:"new_password"  validate:"required,min=6"`
	}

	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("change password failed: body parse", c, err, "userID", userID)
		return HandleError(c, service.ErrInvalidInput)
	}
	if err:= validator.Validate.Struct(req);err!=nil{
		return c.Status(400).JSON(fiber.Map{
			"errors":validator.FormatErrors(err),
		})
	}


	if err := h.authSvc.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		logging.LogWarn("change password failed: service error,auth,Svc problem", c, err, "userID", userID)
		return HandleError(c, err)
	}

	logging.LogInfo("password changed successfully", c, "userID", userID)
	return c.JSON(fiber.Map{"message": "Password changed successfully"})
}


//ForgotPassword
func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.BodyParser(&req); err != nil {
        logging.LogWarn("forgot password failed: body parse", c, err)
		return HandleError(c, service.ErrInvalidInput)
	}
	if err:= validator.Validate.Struct(req);err!=nil{
		return c.Status(400).JSON(fiber.Map{
			"errors":validator.FormatErrors(err),
		})
	}


	err := h.authSvc.ForgotPassword(req.Email)
	if err != nil {
		logging.LogWarn("forgot password failed: service error,auth,Svc problem", c, err)	
		return HandleError(c, err)
	}

	return c.JSON(fiber.Map{
		"message": "If the email exists, an OTP has been sent",
	})
}


// POST /auth/verify-otp
func (h *AuthHandler) VerifyOTP(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email" validate:"required,email"`
		OTP   string `json:"otp" validate:"required,len=6"`
	}

	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("verify OTP failed: body parse", c, err)
		return HandleError(c, service.ErrInvalidInput.WithContext("parsing OTP verification request"))
	}

	if err:= validator.Validate.Struct(req);err!=nil{
		return c.Status(400).JSON(fiber.Map{
			"errors":validator.FormatErrors(err),
		})
	}

	if err := h.authSvc.VerifyOTP(req.Email, req.OTP); err != nil {
		logging.LogWarn("verify OTP failed: service error", c, err, "email", req.Email)
		return HandleError(c, err)
	}

	logging.LogInfo("OTP verified successfully", c, "email", req.Email)
	return c.JSON(fiber.Map{"message": "Account verified successfully"})
}

//ResetPasswordWithOTP
func (h *AuthHandler) ResetPasswordWithOTP(c *fiber.Ctx) error {
	var req struct {
		Email       string `json:"email" validate:"required,email"`
		OTP         string `json:"otp"  validate:"required,len=6"`
		NewPassword string `json:"new_password" validate:"required,min=6"`
	}

	if err := c.BodyParser(&req); err != nil {
		logging.LogWarn("reset password with OTP failed: body parse", c, err)
		return HandleError(c, service.ErrInvalidInput.WithContext("parsing reset password request"))
	}

	if err:= validator.Validate.Struct(req);err!=nil{
		return c.Status(400).JSON(fiber.Map{
			"errors":validator.FormatErrors(err),
		})
	}


	if err := h.authSvc.ResetPassword(req.Email, req.OTP, req.NewPassword); err != nil {
		logging.LogWarn("reset password with OTP failed: service error", c, err, "email", req.Email)
		return HandleError(c, err)
	}

	logging.LogInfo("password reset successfully via OTP", c, "email", req.Email)
	return c.JSON(fiber.Map{"message": "Password reset successfully"})
}

//ResendVerification
func (h *AuthHandler) ResendVerification(c *fiber.Ctx) error {
	type request struct {
		Email string `json:"email"`
	}

	var body request
	if err := c.BodyParser(&body); err != nil {
		logging.LogWarn("resend verification failed: body parse", c, err)
		return fiber.ErrBadRequest
	}

	err := h.authSvc.ResendVerificationEmail(body.Email)
	if err != nil {
		logging.LogWarn("resend verification failed: service error", c, err, "email", body.Email)

		if se, ok := err.(*service.ServiceError); ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code": se.Code,
				"msg":  se.Msg,
			})
		}

		logging.LogWarn("resend verification: unexpected error", c, err, "email", body.Email)
		return fiber.ErrInternalServerError
	}

	logging.LogInfo("resend verification succeeded", c, "email", body.Email)

	return c.JSON(fiber.Map{
		"msg": "Verification OTP resent successfully",
	})
}


func (h *AuthHandler) Logout(c *fiber.Ctx) error {

    // Get refresh token from cookie OR request body
    refreshToken := c.Cookies("refresh_token")
    
    if refreshToken == "" {
        var req struct {
            RefreshToken string `json:"refresh_token"`
        }
        if err := c.BodyParser(&req); err == nil {
            refreshToken = req.RefreshToken
        }
    }
    
    if refreshToken == "" {
        return HandleError(c, service.ErrInvalidInput.WithContext("no refresh token provided"))
    }
    
    if err := h.authSvc.Logout(refreshToken); err != nil {
        return HandleError(c, err)
    }
    
    // Clear cookie
    c.Cookie(&fiber.Cookie{
        Name:     "refresh_token",
        Value:    "",
        Expires:  time.Now().Add(-time.Hour),
        HTTPOnly: true,
        Path:     "/",
    })
    
    logging.LogInfo("user logged out", c)
    return c.JSON(fiber.Map{"message": "logged out"})
}