package app

import (
	"errors"
	"fmt"
	"log"
)

// AppError represents application-specific errors with context
type AppError struct {
	Type    ErrorType
	Message string
	Cause   error
	Context map[string]interface{}
}

// ErrorType categorizes different types of application errors
type ErrorType int

const (
	ErrorTypeAPI ErrorType = iota
	ErrorTypeUI
	ErrorTypeState
	ErrorTypeSystem
	ErrorTypeValidation
	ErrorTypeNetwork
	ErrorTypeFile
)

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error for error chains
func (e *AppError) Unwrap() error {
	return e.Cause
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, message string, cause error) *AppError {
	return &AppError{
		Type:    errorType,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	e.Context[key] = value
	return e
}

// ErrorHandler provides centralized error handling for the application
type ErrorHandler struct {
	logger *log.Logger
	state  *AppState
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logger *log.Logger, state *AppState) *ErrorHandler {
	return &ErrorHandler{
		logger: logger,
		state:  state,
	}
}

// HandleError processes errors and determines appropriate response
func (eh *ErrorHandler) HandleError(err error) ErrorResponse {
	if err == nil {
		return ErrorResponse{ShouldContinue: true}
	}

	eh.logError(err)

	var appErr *AppError
	if errors.As(err, &appErr) {
		return eh.handleAppError(appErr)
	}

	return eh.handleUnknownError(err)
}

// ErrorResponse indicates how the application should respond to an error
type ErrorResponse struct {
	ShouldContinue bool
	ShowModal      bool
	Message        string
	ResetState     bool
}

func (eh *ErrorHandler) handleAppError(err *AppError) ErrorResponse {
	switch err.Type {
	case ErrorTypeAPI:
		return ErrorResponse{
			ShouldContinue: true,
			ShowModal:      true,
			Message:        fmt.Sprintf("API Error: %s", err.Message),
			ResetState:     false,
		}
	case ErrorTypeUI:
		return ErrorResponse{
			ShouldContinue: true,
			ShowModal:      false,
			Message:        "",
			ResetState:     true,
		}
	case ErrorTypeState:
		return ErrorResponse{
			ShouldContinue: true,
			ShowModal:      false,
			Message:        "",
			ResetState:     true,
		}
	case ErrorTypeSystem:
		return ErrorResponse{
			ShouldContinue: true,
			ShowModal:      true,
			Message:        fmt.Sprintf("System Error: %s", err.Message),
			ResetState:     false,
		}
	case ErrorTypeValidation:
		return ErrorResponse{
			ShouldContinue: true,
			ShowModal:      true,
			Message:        fmt.Sprintf("Validation Error: %s", err.Message),
			ResetState:     false,
		}
	case ErrorTypeNetwork:
		return ErrorResponse{
			ShouldContinue: true,
			ShowModal:      true,
			Message:        "Network error. Please check your connection.",
			ResetState:     false,
		}
	case ErrorTypeFile:
		return ErrorResponse{
			ShouldContinue: true,
			ShowModal:      true,
			Message:        fmt.Sprintf("File Error: %s", err.Message),
			ResetState:     false,
		}
	default:
		return eh.handleUnknownError(err)
	}
}

func (eh *ErrorHandler) handleUnknownError(err error) ErrorResponse {
	return ErrorResponse{
		ShouldContinue: true,
		ShowModal:      true,
		Message:        "An unexpected error occurred. Please try again.",
		ResetState:     false,
	}
}

func (eh *ErrorHandler) logError(err error) {
	if eh.logger != nil {
		var appErr *AppError
		if errors.As(err, &appErr) {
			eh.logger.Printf("AppError [%d]: %s", appErr.Type, appErr.Error())
			if len(appErr.Context) > 0 {
				eh.logger.Printf("  Context: %+v", appErr.Context)
			}
		}
	}
}

// Common error creation helpers

func NewAPIError(message string, cause error) *AppError {
	return NewAppError(ErrorTypeAPI, message, cause)
}

func NewUIError(message string, cause error) *AppError {
	return NewAppError(ErrorTypeUI, message, cause)
}

func NewStateError(message string, cause error) *AppError {
	return NewAppError(ErrorTypeState, message, cause)
}

func NewSystemError(message string, cause error) *AppError {
	return NewAppError(ErrorTypeSystem, message, cause)
}

func NewValidationError(message string, cause error) *AppError {
	return NewAppError(ErrorTypeValidation, message, cause)
}

func NewNetworkError(message string, cause error) *AppError {
	return NewAppError(ErrorTypeNetwork, message, cause)
}

func NewFileError(message string, cause error) *AppError {
	return NewAppError(ErrorTypeFile, message, cause)
}

// RecoverFromPanic handles panics and converts them to errors
func RecoverFromPanic() error {
	if r := recover(); r != nil {
		switch v := r.(type) {
		case error:
			return NewAppError(ErrorTypeSystem, "Panic recovered", v)
		case string:
			return NewAppError(ErrorTypeSystem, "Panic recovered", fmt.Errorf(v))
		default:
			return NewAppError(ErrorTypeSystem, "Panic recovered", fmt.Errorf("%v", v))
		}
	}
	return nil
}
