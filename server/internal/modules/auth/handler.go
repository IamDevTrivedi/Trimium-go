package auth

import (
	"errors"
	"net/http"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"server/internal/crypto"
	"server/internal/middleware"
	"server/internal/response"
)

type Handler interface {
	HandleSignUpInit(c *gin.Context)
	HandleSignUpVerify(c *gin.Context)
	HandleCreateUser(c *gin.Context)
	HandleResetPasswordInit(c *gin.Context)
	HandleResetPasswordVerify(c *gin.Context)
	HandleResetPassword(c *gin.Context)
}

type handler struct {
	service Service
}

func NewHandler(s Service) Handler {
	return &handler{service: s}
}

type signUpInitRequest struct {
	Email string `json:"email"`
}

type resetPasswordInitRequest struct {
	Email string `json:"email"`
}

type resetPasswordVerifyRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type resetPasswordRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type signUpVerifyRequest struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type createUserRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

var (
	emailValidator = validator.New()

	errEmailMustBeLowercase = errors.New("email must be lowercase")
	errEmailAliasNotAllowed = errors.New("email aliases are not allowed")
	errEmailInvalid         = errors.New("invalid email address")

	errPasswordTooShort = errors.New("password is too short")
	errPasswordTooLong  = errors.New("password is too long")
	errPasswordTooWeak  = errors.New("password does not meet complexity requirements")
)

func (h *handler) HandleSignUpInit(c *gin.Context) {
	var req signUpInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	email, err := validateSignUpEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusBadRequest, emailValidationMessage(err))
		return
	}

	if err := h.service.SignUpInit(
		c.Request.Context(),
		email,
		middleware.GetStringRequestIP(c),
		middleware.GetStringRequestLocation(c),
		middleware.GetStringRequestUA(c),
	); err != nil {
		if errors.Is(err, ErrEmailAlreadyRegistered) {
			response.Error(c, http.StatusConflict, "Email is already registered")
			return
		}

		response.Error(c, http.StatusInternalServerError, "Failed to initiate signup")
		return
	}

	response.Success(c, "Verification code sent successfully", nil)
}

func (h *handler) HandleSignUpVerify(c *gin.Context) {
	var req signUpVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	email, err := validateSignUpEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusBadRequest, emailValidationMessage(err))
		return
	}

	otp := strings.TrimSpace(req.OTP)
	if err := crypto.ValidateOTP(otp, signUpOTPLength); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid verification code")
		return
	}

	if err := h.service.SignUpVerify(c.Request.Context(), email, otp); err != nil {
		switch {
		case errors.Is(err, ErrSignUpOTPNotFound):
			response.Error(c, http.StatusBadRequest, "Verification code not found or expired")
		case errors.Is(err, ErrInvalidOTP):
			response.Error(c, http.StatusBadRequest, "Invalid verification code")
		case errors.Is(err, ErrMaxVerificationAttempts):
			response.Error(c, http.StatusBadRequest, "Maximum verification attempts exceeded")
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to verify signup")
		}
		return
	}

	response.Success(c, "Email verified successfully", nil)
}

func (h *handler) HandleCreateUser(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	firstName := strings.TrimSpace(req.FirstName)
	if !isValidName(firstName) {
		response.Error(
			c,
			http.StatusBadRequest,
			"First name must be 1-50 characters and contain only letters, spaces, hyphens, or apostrophes",
		)
		return
	}

	lastName := strings.TrimSpace(req.LastName)
	if !isValidName(lastName) {
		response.Error(
			c,
			http.StatusBadRequest,
			"Last name must be 1-50 characters and contain only letters, spaces, hyphens, or apostrophes",
		)
		return
	}

	email, err := validateSignUpEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusBadRequest, emailValidationMessage(err))
		return
	}

	username := strings.TrimSpace(req.Username)
	if !isValidUsername(username) {
		response.Error(
			c,
			http.StatusBadRequest,
			"Username must be 3-30 characters and contain only letters, numbers, hyphens, or periods",
		)
		return
	}

	if err := validatePassword(req.Password); err != nil {
		response.Error(c, http.StatusBadRequest, passwordValidationMessage(err))
		return
	}

	err = h.service.SignUpCreate(
		c.Request.Context(),
		firstName,
		lastName,
		email,
		username,
		req.Password,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrSignUpNotVerified):
			response.Error(c, http.StatusBadRequest, "Email is not verified or verification has expired")
		case errors.Is(err, ErrEmailAlreadyRegistered):
			response.Error(c, http.StatusConflict, "Email is already registered")
		case errors.Is(err, ErrUsernameAlreadyTaken):
			response.Error(c, http.StatusConflict, "Username is already taken")
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to create user")
		}
		return
	}

	response.Created(c, "User account created successfully", nil)
}

func (h *handler) HandleResetPasswordInit(c *gin.Context) {
	var req resetPasswordInitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	email, err := validateSignUpEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusBadRequest, emailValidationMessage(err))
		return
	}

	if err := h.service.ResetPasswordInit(
		c.Request.Context(),
		email,
		middleware.GetStringRequestIP(c),
		middleware.GetStringRequestLocation(c),
		middleware.GetStringRequestUA(c),
	); err != nil {
		if errors.Is(err, ErrEmailNotFound) {
			response.Error(c, http.StatusNotFound, "Email is not registered")
			return
		}

		response.Error(c, http.StatusInternalServerError, "Failed to initiate password reset")
		return
	}

	response.Success(c, "Verification code sent successfully", nil)
}

func (h *handler) HandleResetPasswordVerify(c *gin.Context) {
	var req resetPasswordVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	email, err := validateSignUpEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusBadRequest, emailValidationMessage(err))
		return
	}

	otp := strings.TrimSpace(req.OTP)
	if err := crypto.ValidateOTP(otp, resetPasswordOTPLength); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid verification code")
		return
	}

	if err := h.service.ResetPasswordVerify(c.Request.Context(), email, otp); err != nil {
		switch {
		case errors.Is(err, ErrResetPasswordOTPNotFound):
			response.Error(c, http.StatusBadRequest, "Verification code not found or expired")
		case errors.Is(err, ErrInvalidOTP):
			response.Error(c, http.StatusBadRequest, "Invalid verification code")
		case errors.Is(err, ErrMaxVerificationAttempts):
			response.Error(c, http.StatusBadRequest, "Maximum verification attempts exceeded")
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to verify password reset")
		}
		return
	}

	response.Success(c, "Email verified successfully", nil)
}

func (h *handler) HandleResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	email, err := validateSignUpEmail(req.Email)
	if err != nil {
		response.Error(c, http.StatusBadRequest, emailValidationMessage(err))
		return
	}

	if err := validatePassword(req.Password); err != nil {
		response.Error(c, http.StatusBadRequest, passwordValidationMessage(err))
		return
	}

	err = h.service.ResetPassword(c.Request.Context(), email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, ErrResetPasswordNotVerified):
			response.Error(c, http.StatusBadRequest, "Password reset is not verified or verification has expired")
		case errors.Is(err, ErrEmailNotFound):
			response.Error(c, http.StatusNotFound, "Email is not registered")
		default:
			response.Error(c, http.StatusInternalServerError, "Failed to reset password")
		}
		return
	}

	response.Success(c, "Password reset successfully", nil)
}

func validateSignUpEmail(raw string) (string, error) {
	email := strings.TrimSpace(raw)

	if email != strings.ToLower(email) {
		return "", errEmailMustBeLowercase
	}

	if strings.Contains(email, "+") {
		return "", errEmailAliasNotAllowed
	}

	if err := emailValidator.Var(email, "required,email"); err != nil {
		return "", errEmailInvalid
	}

	return email, nil
}

func isValidName(name string) bool {
	length := utf8.RuneCountInString(name)
	if length < minNameLength || length > maxNameLength {
		return false
	}

	for _, r := range name {
		switch {
		case unicode.IsLetter(r), unicode.IsMark(r):
		case r == ' ', r == '-', r == '\'':
		default:
			return false
		}
	}

	return true
}

func isValidUsername(username string) bool {
	length := utf8.RuneCountInString(username)
	if length < minUsernameLength || length > maxUsernameLength {
		return false
	}

	for _, r := range username {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-', r == '.':
		default:
			return false
		}
	}

	return true
}

func validatePassword(password string) error {
	length := utf8.RuneCountInString(password)
	if length < minPasswordLength {
		return errPasswordTooShort
	}
	if length > maxPasswordLength {
		return errPasswordTooLong
	}

	var hasLower, hasUpper, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsSpace(r):
			hasSpecial = true
		}
	}

	if !hasLower || !hasUpper || !hasDigit || !hasSpecial {
		return errPasswordTooWeak
	}

	return nil
}

func emailValidationMessage(err error) string {
	switch {
	case errors.Is(err, errEmailMustBeLowercase):
		return "Email must be lowercase"
	case errors.Is(err, errEmailAliasNotAllowed):
		return "Email aliases are not allowed"
	default:
		return "Invalid email address"
	}
}

func passwordValidationMessage(err error) string {
	switch {
	case errors.Is(err, errPasswordTooShort):
		return "Password must be at least 10 characters"
	case errors.Is(err, errPasswordTooLong):
		return "Password must be at most 128 characters"
	default:
		return "Password must contain uppercase, lowercase, number, and special character"
	}
}
