package middleware

import (
	"fmt"
	"regexp"
	"strings"
)

var spaceReg, _ = regexp.Compile("\\s{2,}")

func CompressSpace(str string) string {
	return spaceReg.ReplaceAllString(str, " ")
}

func ReplaceStr(ori string, old string, newFunc func() string) string {
	for {
		res := strings.Replace(ori, old, newFunc(), -1)
		if ori == res {
			return ori
		} else {
			ori = res
		}
	}
}

const TimeFormat = "2006-1-2 15:04:05"

const (
	defaultServerName  = "middleware"
	defaultUserAgent   = "middleware"
	defaultContentType = "text/html; charset=utf-8"
)

const (
	Slash               = "/"
	SlashSlash          = "//"
	SlashDotDot         = "/.."
	SlashDotSlash       = "/./"
	SlashDotDotSlash    = "/../"
	CRLF                = "\r\n"
	HTTP                = "http"
	HTTPS               = "https"
	HTTP11              = "HTTP/1.1"
	ColonSlashSlash     = "://"
	ColonSpace          = ": "
	GMT                 = "GMT"
	ResponseContinue    = "HTTP/1.1 100 Continue\r\n\r\n"
	GET                 = "GET"
	HEAD                = "HEAD"
	POST                = "POST"
	PUT                 = "PUT"
	DELETE              = "DELETE"
	OPTIONS             = "OPTIONS"
	EXPECT              = "EXPECT"
	Connection          = "Connection"
	ContentLength       = "Content-Length"
	ContentType         = "Content-Type"
	Date                = "Date"
	Host                = "Host"
	Referer             = "Referer"
	ServerHeader        = "Server"
	TransferEncoding    = "Transfer-Encoding"
	ContentEncoding     = "Content-Encoding"
	AcceptEncoding      = "Accept-Encoding"
	UserAgent           = "User-Agent"
	Cookie              = "Cookie"
	SetCookie           = "Set-Cookie"
	Location            = "Location"
	IfModifiedSince     = "If-Modified-Since"
	LastModified        = "Last-Modified"
	AcceptRanges        = "Accept-Ranges"
	Range               = "Range"
	ContentRange        = "Content-Range"
	CookieExpires       = "expires"
	CookieDomain        = "domain"
	CookiePath          = "Path"
	CookieHTTPOnly      = "HttpOnly"
	CookieSecure        = "secure"
	HttpClose           = "close"
	Gzip                = "gzip"
	Deflate             = "deflate"
	KeepAlive           = "keep-alive"
	KeepAliveCamelCase  = "Keep-Alive"
	Upgrade             = "Upgrade"
	Chunked             = "chunked"
	Identity            = "identity"
	PostArgsContentType = "application/x-www-form-urlencoded"
	MultipartFormData   = "multipart/form-data"
	Boundary            = "boundary"
	Bytes               = "bytes"
	TextSlash           = "text/"
	ApplicationSlash    = "application/"
)

const (
	ApplicationJson   = "application/json; charset=utf-8"
	Css               = "text/css; charset=utf-8"
	Plain             = "text/plain; charset=utf-8"
	Html              = "text/html; charset=utf-8"
	Jpeg              = "image/jpeg"
	Js                = "application/x-javascript; charset=utf-8"
	Pdf               = "application/pdf"
	Png               = "image/png"
	Svg               = "image/svg+xml"
	Xml               = "text/xml; charset=utf-8"
	ApplicationFont   = "application/x-font-woff"
	ApplicationStream = "application/octet-stream"
)

const (
	AccessControlAllowOrigin  = "Access-Control-Allow-Origin"
	AccessControlAllowMethods = "Access-Control-Allow-Methods"
	AccessControlAllowHeaders = "Access-Control-Allow-Headers"
	METHODS                   = "POST,GET,OPTIONS,DELETE"
)

// HTTP status codes as registered with IANA.
// See: http://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml
const (
	StatusContinue           = 100 // RFC 7231, 6.2.1
	StatusSwitchingProtocols = 101 // RFC 7231, 6.2.2
	StatusProcessing         = 102 // RFC 2518, 10.1

	StatusOK                   = 200 // RFC 7231, 6.3.1
	StatusCreated              = 201 // RFC 7231, 6.3.2
	StatusAccepted             = 202 // RFC 7231, 6.3.3
	StatusNonAuthoritativeInfo = 203 // RFC 7231, 6.3.4
	StatusNoContent            = 204 // RFC 7231, 6.3.5
	StatusResetContent         = 205 // RFC 7231, 6.3.6
	StatusPartialContent       = 206 // RFC 7233, 4.1
	StatusMultiStatus          = 207 // RFC 4918, 11.1
	StatusAlreadyReported      = 208 // RFC 5842, 7.1
	StatusIMUsed               = 226 // RFC 3229, 10.4.1

	StatusMultipleChoices   = 300 // RFC 7231, 6.4.1
	StatusMovedPermanently  = 301 // RFC 7231, 6.4.2
	StatusFound             = 302 // RFC 7231, 6.4.3
	StatusSeeOther          = 303 // RFC 7231, 6.4.4
	StatusNotModified       = 304 // RFC 7232, 4.1
	StatusUseProxy          = 305 // RFC 7231, 6.4.5
	_                       = 306 // RFC 7231, 6.4.6 (Unused)
	StatusTemporaryRedirect = 307 // RFC 7231, 6.4.7
	StatusPermanentRedirect = 308 // RFC 7538, 3

	StatusBadRequest                   = 400 // RFC 7231, 6.5.1
	StatusUnauthorized                 = 401 // RFC 7235, 3.1
	StatusPaymentRequired              = 402 // RFC 7231, 6.5.2
	StatusForbidden                    = 403 // RFC 7231, 6.5.3
	StatusNotFound                     = 404 // RFC 7231, 6.5.4
	StatusMethodNotAllowed             = 405 // RFC 7231, 6.5.5
	StatusNotAcceptable                = 406 // RFC 7231, 6.5.6
	StatusProxyAuthRequired            = 407 // RFC 7235, 3.2
	StatusRequestTimeout               = 408 // RFC 7231, 6.5.7
	StatusConflict                     = 409 // RFC 7231, 6.5.8
	StatusGone                         = 410 // RFC 7231, 6.5.9
	StatusLengthRequired               = 411 // RFC 7231, 6.5.10
	StatusPreconditionFailed           = 412 // RFC 7232, 4.2
	StatusRequestEntityTooLarge        = 413 // RFC 7231, 6.5.11
	StatusRequestURITooLong            = 414 // RFC 7231, 6.5.12
	StatusUnsupportedMediaType         = 415 // RFC 7231, 6.5.13
	StatusRequestedRangeNotSatisfiable = 416 // RFC 7233, 4.4
	StatusExpectationFailed            = 417 // RFC 7231, 6.5.14
	StatusTeapot                       = 418 // RFC 7168, 2.3.3
	StatusUnprocessableEntity          = 422 // RFC 4918, 11.2
	StatusLocked                       = 423 // RFC 4918, 11.3
	StatusFailedDependency             = 424 // RFC 4918, 11.4
	StatusUpgradeRequired              = 426 // RFC 7231, 6.5.15
	StatusPreconditionRequired         = 428 // RFC 6585, 3
	StatusTooManyRequests              = 429 // RFC 6585, 4
	StatusRequestHeaderFieldsTooLarge  = 431 // RFC 6585, 5
	StatusUnavailableForLegalReasons   = 451 // RFC 7725, 3

	StatusInternalServerError           = 500 // RFC 7231, 6.6.1
	StatusNotImplemented                = 501 // RFC 7231, 6.6.2
	StatusBadGateway                    = 502 // RFC 7231, 6.6.3
	StatusServiceUnavailable            = 503 // RFC 7231, 6.6.4
	StatusGatewayTimeout                = 504 // RFC 7231, 6.6.5
	StatusHTTPVersionNotSupported       = 505 // RFC 7231, 6.6.6
	StatusVariantAlsoNegotiates         = 506 // RFC 2295, 8.1
	StatusInsufficientStorage           = 507 // RFC 4918, 11.5
	StatusLoopDetected                  = 508 // RFC 5842, 7.2
	StatusNotExtended                   = 510 // RFC 2774, 7
	StatusNetworkAuthenticationRequired = 511 // RFC 6585, 6
)

var statusText = map[int]string{
	StatusContinue:           "Continue",
	StatusSwitchingProtocols: "Switching Protocols",
	StatusProcessing:         "Processing",

	StatusOK:                   "OK",
	StatusCreated:              "Created",
	StatusAccepted:             "Accepted",
	StatusNonAuthoritativeInfo: "Non-Authoritative Information",
	StatusNoContent:            "No Content",
	StatusResetContent:         "Reset Content",
	StatusPartialContent:       "Partial Content",
	StatusMultiStatus:          "Multi-Status",
	StatusAlreadyReported:      "Already Reported",
	StatusIMUsed:               "IM Used",

	StatusMultipleChoices:   "Multiple Choices",
	StatusMovedPermanently:  "Moved Permanently",
	StatusFound:             "Found",
	StatusSeeOther:          "See Other",
	StatusNotModified:       "Not Modified",
	StatusUseProxy:          "Use Proxy",
	StatusTemporaryRedirect: "Temporary Redirect",
	StatusPermanentRedirect: "Permanent Redirect",

	StatusBadRequest:                   "Bad Request",
	StatusUnauthorized:                 "Unauthorized",
	StatusPaymentRequired:              "Payment Required",
	StatusForbidden:                    "Forbidden",
	StatusNotFound:                     "Not Found",
	StatusMethodNotAllowed:             "Method Not Allowed",
	StatusNotAcceptable:                "Not Acceptable",
	StatusProxyAuthRequired:            "Proxy Authentication Required",
	StatusRequestTimeout:               "Request Timeout",
	StatusConflict:                     "Conflict",
	StatusGone:                         "Gone",
	StatusLengthRequired:               "Length Required",
	StatusPreconditionFailed:           "Precondition Failed",
	StatusRequestEntityTooLarge:        "Request Entity Too Large",
	StatusRequestURITooLong:            "Request URI Too Long",
	StatusUnsupportedMediaType:         "Unsupported Media Type",
	StatusRequestedRangeNotSatisfiable: "Requested Range Not Satisfiable",
	StatusExpectationFailed:            "Expectation Failed",
	StatusTeapot:                       "I'm a teapot",
	StatusUnprocessableEntity:          "Unprocessable Entity",
	StatusLocked:                       "Locked",
	StatusFailedDependency:             "Failed Dependency",
	StatusUpgradeRequired:              "Upgrade Required",
	StatusPreconditionRequired:         "Precondition Required",
	StatusTooManyRequests:              "Too Many Requests",
	StatusRequestHeaderFieldsTooLarge:  "Request Header Fields Too Large",
	StatusUnavailableForLegalReasons:   "Unavailable For Legal Reasons",

	StatusInternalServerError:           "Internal Server Error",
	StatusNotImplemented:                "Not Implemented",
	StatusBadGateway:                    "Bad Gateway",
	StatusServiceUnavailable:            "Service Unavailable",
	StatusGatewayTimeout:                "Gateway Timeout",
	StatusHTTPVersionNotSupported:       "HTTP Version Not Supported",
	StatusVariantAlsoNegotiates:         "Variant Also Negotiates",
	StatusInsufficientStorage:           "Insufficient Storage",
	StatusLoopDetected:                  "Loop Detected",
	StatusNotExtended:                   "Not Extended",
	StatusNetworkAuthenticationRequired: "Network Authentication Required",
}

// StatusText returns a text for the HTTP status code. It returns the empty
// string if the code is unknown.
func StatusText(code int) string {
	return statusText[code]
}

const StaticHtml = `<!DOCTYPE html>
<html lang="en">

<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">

    <title>%v</title>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <style>
        * {
            line-height: 1.2;
            margin: 0;
        }

        html {
            color: #888;
            display: table;
            font-family: sans-serif;
            height: 100%%;
            text-align: center;
            width: 100%%;
        }

        body {
            display: table-cell;
            vertical-align: middle;
            margin: 2em auto;
        }

        h1 {
            color: #555;
            font-size: 2em;
            font-weight: 400;
        }

        p {
            margin: 0 auto;
            width: 280px;
        }

        @media only screen and (max-width: 280px) {

            body,
            p {
                width: 95%%;
            }

            h1 {
                font-size: 1.5em;
                margin: 0 0 0.3em;
            }

        }
    </style>
</head>
<body>
    %v
</body>
</html>
`

var StatusNotFoundView = fmt.Sprintf(StaticHtml, "NOT FOUND", "<h1>404 NOT FOUND</h1>")

const (
	red   = "\033[31m"
	green = "\033[32m"
	end   = "\033[0m"
)

func ColorPrint(out string, color string) {
	switch color {
	case "green":
		fmt.Printf("%s%s%s\n", green, out, end)
		break
	case "red":
		fmt.Printf("%s%s%s\n", red, out, end)
		break
	}
}
