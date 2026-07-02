package apperrors

type Code string

const (
	CodeInternalError Code = "INTERNAL_ERROR"

	CodeInvalidRequestBody      Code = "INVALID_REQUEST_BODY"
	CodeValidationRequired      Code = "VALIDATION_REQUIRED_FIELDS"
	CodeValidationPasswordShort Code = "VALIDATION_PASSWORD_TOO_SHORT"
	CodeValidationInvalidRole   Code = "VALIDATION_INVALID_ROLE"
	CodeValidationNoFields      Code = "VALIDATION_NO_FIELDS"

	CodeUnauthorized          Code = "UNAUTHORIZED"
	CodeForbidden             Code = "FORBIDDEN"
	CodeAuthInvalidHeader     Code = "AUTH_INVALID_HEADER"
	CodeAuthInvalidToken      Code = "AUTH_INVALID_TOKEN"
	CodeAuthInvalidCredentials Code = "AUTH_INVALID_CREDENTIALS"
	CodeAuthRefreshMissing    Code = "AUTH_REFRESH_MISSING"
	CodeAuthRefreshInvalid    Code = "AUTH_REFRESH_INVALID"

	CodeUserNotFound      Code = "USER_NOT_FOUND"
	CodeUserEmailTaken    Code = "USER_EMAIL_TAKEN"
	CodeUserLastAdmin     Code = "USER_LAST_ADMIN"
	CodeUserCreateFailed  Code = "USER_CREATE_FAILED"
	CodeUserUpdateFailed  Code = "USER_UPDATE_FAILED"
	CodeUserDeleteFailed  Code = "USER_DELETE_FAILED"
	CodeUserListFailed    Code = "USER_LIST_FAILED"
	CodeUserGetFailed     Code = "USER_GET_FAILED"

	CodeTokenCreateFailed         Code = "TOKEN_CREATE_FAILED"
	CodeRefreshTokenCreateFailed  Code = "REFRESH_TOKEN_CREATE_FAILED"
	CodeRefreshTokenStoreFailed   Code = "REFRESH_TOKEN_STORE_FAILED"
	CodeRefreshTokenValidateFailed Code = "REFRESH_TOKEN_VALIDATE_FAILED"
	CodeRefreshTokenRotateFailed  Code = "REFRESH_TOKEN_ROTATE_FAILED"
	CodeRefreshFailed             Code = "REFRESH_FAILED"
	CodeAuthenticateFailed        Code = "AUTHENTICATE_FAILED"
	CodePasswordHashFailed        Code = "PASSWORD_HASH_FAILED"
)
