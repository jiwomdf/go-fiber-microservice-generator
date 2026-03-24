package constant

import "errors"

var ErrQueryError = errors.New("query_error")

const (
	MaxUint32 = ^uint32(0)
	MinUint32 = 0
	MaxInt32  = int(MaxUint32 >> 1)
	MinInt32  = -MaxInt32 - 1
	MaxUint64 = ^uint64(0)
	MinUint64 = 0
	MaxInt64  = int(MaxUint64 >> 1)
	MinInt64  = -MaxInt64 - 1
)

const (
	MimeTypeImagePng  = "image/png"
	MimeTypeImageJpeg = "image/jpeg"
	MimeTypeImageWebp = "image/webp"
	MimeTypePdf       = "application/pdf"
	MimeTypeXPdf      = "application/x-pdf"
)

const (
	DateTimeFormat         = "2006-01-02T15:04:05+07:00"
	MaxSizeBanner          = 10 * 1024 * 1024
	MaxSizeHotelLogo       = 5 * 1024 * 1024
	MaxSizeHotelBackground = 15 * 1024 * 1024
	TimeZoneAsiaJakarta    = "Asia/Jakarta"
)

const (
	MimeTypeVideoMp4        = "video/mp4"
	MimeTypeVideoMpd        = "application/dash+xml"
	MimeTypeVideoM3u8       = "application/vnd.apple.mpegurl"
	MimeTypeVideoM3u8Legacy = "application/x-mpegURL"
)
