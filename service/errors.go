package service

import "fmt"

type ServiceError struct {
	Code string
	Msg  string
}

func (e *ServiceError) Error() string {
	return e.Msg
}

// Optional
func (e *ServiceError) WithContext(ctx string) *ServiceError {
	return &ServiceError{
		Code: e.Code,
		Msg:  fmt.Sprintf("%s: %s", e.Msg, ctx),
	}
}

// Predefined service errors
var (
	ErrUserNotFound      = &ServiceError{Code: "USER_NOT_FOUND", Msg: "user not found"}
	ErrUserBlocked       = &ServiceError{Code: "USER_BLOCKED", Msg: "user is blocked"}
	ErrForbidden         = &ServiceError{Code: "FORBIDDEN_ACTION", Msg: "forbidden action"}
	ErrInvalidInput      = &ServiceError{Code: "INVALID_INPUT", Msg: "invalid input"}
	ErrUserExists        = &ServiceError{Code: "USER_EXISTS", Msg: "user with this email already exists"}
	ErrInvalidLogin      = &ServiceError{Code: "INVALID_LOGIN", Msg: "invalid email or password"}
	ErrOTPInvalid        = &ServiceError{Code: "OTP_INVALID", Msg: "invalid or expired OTP"}
	ErrInvalidToken      = &ServiceError{Code: "TOKEN_INVALID", Msg: "invalid or expired TOKEN"}
	ErrPasswordMismatch  = &ServiceError{Code: "PASSWORD_MISMATCH", Msg: "current password does not match"}
	ErrWeakPassword      = &ServiceError{Code: "WEAK_PASSWORD", Msg: "password is too weak"}
	ErrPasswordReUse      = &ServiceError{Code: "PASSWORD_REUSE", Msg:  "new password cannot be the same as the old password"}
	// ORDER ERRORS
	ErrOrderNotCancelable = &ServiceError{
		Code: "ORDER_NOT_CANCELABLE",
		Msg:  "order cannot be canceled in its current state",
	}
		ErrInvalidOrderStatus = &ServiceError{
		Code: "INVALID_ORDER_STATUS",
		Msg:  "invalid order status",
	}
	// PRODUCT
	ErrProductNotFound = &ServiceError{
		Code: "PRODUCT_NOT_FOUND",
		Msg:  "product not found",
	}

	
	ErrEmailMismatch = &ServiceError{
		Code: "EMAIL_MISMATCH",
		Msg:  "email does not match token",
	}

	ErrEmailNotVerified = &ServiceError{
		Code: "EMAIL_NOT_VERIFIED",
		Msg:  "email is not verified",
	}

	ErrOTPExpired = &ServiceError{
		Code: "OTP_EXPIRED",
		Msg:  "OTP has expired",
	}

	ErrInvalidOTP = &ServiceError{
		Code: "OTP_INVALID",
		Msg:  "invalid OTP",
	}

	ErrAlreadyVerified = &ServiceError{
		Code: "ALREADY_VERIFIED",
		Msg:  "email is already verified",
	}

    ErrUnauthorized = &ServiceError{
        Code: "UNAUTHORIZED",
        Msg:  "unauthorized",
    }
	ErrNotFound = &ServiceError{
		Code: "RESOURCE_NOT_FOUND",
		Msg:  "resource not found",
	}

)


