package apperrors

var messagesEN = map[Code]string{
	CodeInternalError: "An unexpected error occurred. Please try again later.",

	CodeInvalidRequestBody:      "The request body is invalid.",
	CodeValidationRequired:      "Required fields are missing.",
	CodeValidationPasswordShort:   "Password must be at least 8 characters.",
	CodeValidationInvalidRole:   "Role must be ADMIN or USER.",
	CodeValidationNoFields:      "At least one field is required.",

	CodeUnauthorized:           "Authentication is required.",
	CodeForbidden:              "You do not have permission to perform this action.",
	CodeAuthInvalidHeader:      "A valid authorization header is required.",
	CodeAuthInvalidToken:       "The access token is invalid or expired.",
	CodeAuthInvalidCredentials: "The email or password is incorrect.",
	CodeAuthRefreshMissing:     "Refresh token is missing.",
	CodeAuthRefreshInvalid:     "The refresh token is invalid or expired.",

	CodeUserNotFound:     "User not found.",
	CodeUserEmailTaken:   "This email is already in use.",
	CodeUserLastAdmin:    "Cannot modify or remove the last admin.",
	CodeUserCreateFailed: "Failed to create user.",
	CodeUserUpdateFailed: "Failed to update user.",
	CodeUserDeleteFailed: "Failed to delete user.",
	CodeUserListFailed:   "Failed to list users.",
	CodeUserGetFailed:    "Failed to get user.",

	CodeTokenCreateFailed:          "Failed to create access token.",
	CodeRefreshTokenCreateFailed:   "Failed to create refresh token.",
	CodeRefreshTokenStoreFailed:    "Failed to store refresh token.",
	CodeRefreshTokenValidateFailed: "Failed to validate refresh token.",
	CodeRefreshTokenRotateFailed:   "Failed to rotate refresh token.",
	CodeRefreshFailed:              "Failed to refresh access token.",
	CodeAuthenticateFailed:         "Failed to authenticate.",
	CodePasswordHashFailed:         "Failed to process password.",
}

func messageEN(code Code) string {
	if msg, ok := messagesEN[code]; ok {
		return msg
	}
	return messagesEN[CodeInternalError]
}
