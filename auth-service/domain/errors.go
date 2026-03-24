package domain

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc/codes"
)

const StatusCodePrefix = "AUTH"

var (
	StatusSuccess                       = StatusCodePrefix + "20000"
	StatusSuccessCreate                 = StatusCodePrefix + "20100"
	StatusSuccessLogin                  = StatusCodePrefix + "20001"
	StatusSuccessCreateAuth             = StatusCodePrefix + "20101"
	StatusSuccessCreateProperty         = StatusCodePrefix + "20102"
	StatusBadRequest                    = StatusCodePrefix + "40000"
	StatusMissingParameter              = StatusCodePrefix + "40001"
	StatusWrongValue                    = StatusCodePrefix + "40002"
	StatusVendorIdNotExist              = StatusCodePrefix + "40003"
	StatusUnrecognizedLanguageCode      = StatusCodePrefix + "40004"
	StatusUnauthorized                  = StatusCodePrefix + "40100"
	StatusForbidden                     = StatusCodePrefix + "40300"
	StatusForbiddenUpdateStatusCategory = StatusCodePrefix + "40301"
	StatusForbiddenUpdateStatusItem     = StatusCodePrefix + "40302"
	StatusForbiddenDeleteItem           = StatusCodePrefix + "40303"
	StatusNotFound                      = StatusCodePrefix + "40400"
	StatusGuestNotFound                 = StatusCodePrefix + "40401"
	StatusRoomNotFound                  = StatusCodePrefix + "40402"
	StatusPaymentNotFound               = StatusCodePrefix + "40403"
	StatusPaymentQRNotFound             = StatusCodePrefix + "40404"
	StatusHotelNotFound                 = StatusCodePrefix + "40405"
	StatusItemsNotAvailable             = StatusCodePrefix + "40050"
	StatusInternalServerError           = StatusCodePrefix + "50000"
	StatusForbiddenWrongHotelID         = StatusCodePrefix + "40304"
	StatusConflict                      = StatusCodePrefix + "40900"
	StatusWordsAlreadyExist             = StatusCodePrefix + "40901"
	StatusInvalidEmailPassword          = StatusCodePrefix + "40051"
)

var (
	ErrBadRequest          = errors.New("bad request")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrNotFound            = errors.New("not found")
	ErrAuthentication      = errors.New("authentication failed")
	ErrInternalServerError = errors.New("internal server error")
)

func GetHttpStatusCode(status string) int {
	prefixLen := len(StatusCodePrefix)
	// A valid code should be at least prefix + 3 digits for HTTP status
	if len(status) < prefixLen+3 || status[:prefixLen] != StatusCodePrefix {
		return fiber.StatusInternalServerError
	}

	// Extract the HTTP status part, e.g., "404" from "ID40401"
	httpStatusCodeStr := status[prefixLen : prefixLen+3]
	httpStatusCode, err := strconv.Atoi(httpStatusCodeStr)
	if err != nil {
		return fiber.StatusInternalServerError
	}

	// Basic validation of the parsed HTTP status code
	if httpStatusCode < 100 || httpStatusCode > 599 {
		return fiber.StatusInternalServerError
	}

	return httpStatusCode
}

func GetCustomStatusMessage(status string, m string) string {
	switch status {
	case StatusSuccess:
		return "Success"
	case StatusSuccessCreate:
		return "Success create " + m
	case StatusSuccessCreateAuth:
		return "Auth created successfully. A verification email has been sent."
	case StatusSuccessCreateProperty:
		return "Property created successfully."
	case StatusBadRequest:
		return "Bad request"
	case StatusMissingParameter:
		return "Missing parameter: " + m
	case StatusUnauthorized:
		return "Unauthorized"
	case StatusForbidden:
		return "Forbidden"
	case StatusNotFound:
		if m != "" {
			return "Not found: " + m
		}
		return "Not found"
	case StatusWrongValue:
		return "Wrong value for parameter " + m
	default:
		return "Internal server error"
	}
}

func GetStatusFromErr(err error) string {
	switch err {
	case ErrBadRequest:
		return StatusBadRequest
	case ErrUnauthorized:
		return StatusUnauthorized
	case ErrForbidden:
		return StatusForbidden
	case ErrNotFound:
		return StatusNotFound
	default:
		return StatusInternalServerError
	}
}

func GetStatusMessage(err error) string {
	switch err {
	case ErrBadRequest:
		return "Bad request"
	case ErrUnauthorized:
		return "Unauthorized"
	case ErrForbidden:
		return "Forbidden"
	case ErrNotFound:
		return "Not found"
	default:
		return "Internal server error"
	}
}

func GetStatusGRPCErr(cd codes.Code) string {
	switch cd {
	case codes.OK:
		return StatusSuccess
	case codes.InvalidArgument:
		return StatusBadRequest
	case codes.PermissionDenied:
		return StatusForbidden
	case codes.Unauthenticated:
		return StatusUnauthorized
	case codes.NotFound:
		return StatusNotFound
	default:
		return StatusInternalServerError
	}
}
