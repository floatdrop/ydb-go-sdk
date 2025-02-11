package errors

import (
	"bytes"
	"errors"
	"strconv"
	"strings"

	grpc "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Issue"
)

// Issue struct
type Issue struct {
	Message  string
	Code     uint32
	Severity uint32
}

var ErrOperationNotReady = errors.New("operation is not ready yet")

type IssueIterator []*Ydb_Issue.IssueMessage

func (it IssueIterator) Len() int {
	return len(it)
}

func (it IssueIterator) Get(i int) (issue Issue, nested IssueIterator) {
	x := it[i]
	if xs := x.Issues; len(xs) > 0 {
		nested = IssueIterator(xs)
	}
	return Issue{
		Message:  x.GetMessage(),
		Code:     x.GetIssueCode(),
		Severity: x.GetSeverity(),
	}, nested
}

type TransportError struct {
	Reason TransportErrorCode

	message string
	err     error
}

type teOpt func(te *TransportError)

func WithTEReason(reason TransportErrorCode) teOpt {
	return func(te *TransportError) {
		te.Reason = reason
	}
}

// WithTEError stores err into transport error
func WithTEError(err error) teOpt {
	return func(te *TransportError) {
		te.err = err
	}
}

// WithTEMessage stores message into transport error
func WithTEMessage(message string) teOpt {
	return func(te *TransportError) {
		te.message = message
	}
}

// WithTEOperation stores reason code into transport error from operation
func WithTEOperation(operation operation) teOpt {
	return func(te *TransportError) {
		te.Reason = TransportErrorCode(operation.GetStatus())
	}
}

// NewTransportError returns a new transport error with given options
func NewTransportError(opts ...teOpt) error {
	te := &TransportError{
		Reason: TransportErrorUnknownCode,
	}
	for _, f := range opts {
		f(te)
	}
	return te
}

func (t *TransportError) Error() string {
	s := "ydb: transport error: " + t.Reason.String()
	if t.message != "" {
		s += ": " + t.message
	}
	return s
}

func (t *TransportError) Unwrap() error {
	return t.err
}

func dumpIssues(buf *bytes.Buffer, ms []*Ydb_Issue.IssueMessage) {
	if len(ms) == 0 {
		return
	}
	buf.WriteByte(' ')
	buf.WriteByte('[')
	defer buf.WriteByte(']')
	for _, m := range ms {
		buf.WriteByte('{')
		if code := m.GetIssueCode(); code != 0 {
			buf.WriteByte('#')
			buf.WriteString(strconv.Itoa(int(code)))
			buf.WriteByte(' ')
		}
		buf.WriteString(strings.TrimSuffix(m.GetMessage(), "."))
		dumpIssues(buf, m.Issues)
		buf.WriteByte('}')
	}
}

// StatusCode reports unsuccessful operation status code.
type StatusCode int32

func (e StatusCode) String() string {
	return Ydb.StatusIds_StatusCode_name[int32(e)]
}

// Errors describing unsusccessful operation status.
const (
	StatusUnknownStatus      = StatusCode(Ydb.StatusIds_STATUS_CODE_UNSPECIFIED)
	StatusBadRequest         = StatusCode(Ydb.StatusIds_BAD_REQUEST)
	StatusUnauthorized       = StatusCode(Ydb.StatusIds_UNAUTHORIZED)
	StatusInternalError      = StatusCode(Ydb.StatusIds_INTERNAL_ERROR)
	StatusAborted            = StatusCode(Ydb.StatusIds_ABORTED)
	StatusUnavailable        = StatusCode(Ydb.StatusIds_UNAVAILABLE)
	StatusOverloaded         = StatusCode(Ydb.StatusIds_OVERLOADED)
	StatusSchemeError        = StatusCode(Ydb.StatusIds_SCHEME_ERROR)
	StatusGenericError       = StatusCode(Ydb.StatusIds_GENERIC_ERROR)
	StatusTimeout            = StatusCode(Ydb.StatusIds_TIMEOUT)
	StatusBadSession         = StatusCode(Ydb.StatusIds_BAD_SESSION)
	StatusPreconditionFailed = StatusCode(Ydb.StatusIds_PRECONDITION_FAILED)
	StatusAlreadyExists      = StatusCode(Ydb.StatusIds_ALREADY_EXISTS)
	StatusNotFound           = StatusCode(Ydb.StatusIds_NOT_FOUND)
	StatusSessionExpired     = StatusCode(Ydb.StatusIds_SESSION_EXPIRED)
	StatusCancelled          = StatusCode(Ydb.StatusIds_CANCELLED)
	StatusUndetermined       = StatusCode(Ydb.StatusIds_UNDETERMINED)
	StatusUnsupported        = StatusCode(Ydb.StatusIds_UNSUPPORTED)
	StatusSessionBusy        = StatusCode(Ydb.StatusIds_SESSION_BUSY)
)

func statusCode(s Ydb.StatusIds_StatusCode) StatusCode {
	switch s {
	case Ydb.StatusIds_BAD_REQUEST:
		return StatusBadRequest
	case Ydb.StatusIds_UNAUTHORIZED:
		return StatusUnauthorized
	case Ydb.StatusIds_INTERNAL_ERROR:
		return StatusInternalError
	case Ydb.StatusIds_ABORTED:
		return StatusAborted
	case Ydb.StatusIds_UNAVAILABLE:
		return StatusUnavailable
	case Ydb.StatusIds_OVERLOADED:
		return StatusOverloaded
	case Ydb.StatusIds_SCHEME_ERROR:
		return StatusSchemeError
	case Ydb.StatusIds_GENERIC_ERROR:
		return StatusGenericError
	case Ydb.StatusIds_TIMEOUT:
		return StatusTimeout
	case Ydb.StatusIds_BAD_SESSION:
		return StatusBadSession
	case Ydb.StatusIds_PRECONDITION_FAILED:
		return StatusPreconditionFailed
	case Ydb.StatusIds_ALREADY_EXISTS:
		return StatusAlreadyExists
	case Ydb.StatusIds_NOT_FOUND:
		return StatusNotFound
	case Ydb.StatusIds_SESSION_EXPIRED:
		return StatusSessionExpired
	case Ydb.StatusIds_CANCELLED:
		return StatusCancelled
	case Ydb.StatusIds_UNDETERMINED:
		return StatusUndetermined
	case Ydb.StatusIds_UNSUPPORTED:
		return StatusUnsupported
	case Ydb.StatusIds_SESSION_BUSY:
		return StatusSessionBusy
	default:
		return StatusUnknownStatus
	}
}

type TransportErrorCode int32

func (t TransportErrorCode) String() string {
	return transportErrorString(t)
}

func (t TransportErrorCode) OperationCompleted() OperationCompleted {
	switch t {
	case
		TransportErrorAborted,
		TransportErrorResourceExhausted:
		return OperationCompletedFalse
	case
		TransportErrorInternal,
		TransportErrorCanceled,
		TransportErrorUnavailable:
		return OperationCompletedUndefined
	default:
		return OperationCompletedTrue
	}
}

func (t TransportErrorCode) BackoffType() BackoffType {
	switch t {
	case
		TransportErrorInternal,
		TransportErrorCanceled,
		TransportErrorUnavailable:
		return BackoffTypeFastBackoff
	case
		TransportErrorResourceExhausted:
		return BackoffTypeSlowBackoff
	default:
		return BackoffTypeNoBackoff
	}
}

func (t TransportErrorCode) MustDeleteSession() bool {
	switch t {
	case
		TransportErrorResourceExhausted,
		TransportErrorOutOfRange:
		return false
	default:
		return true
	}
}

const (
	TransportErrorUnknownCode TransportErrorCode = iota
	TransportErrorCanceled
	TransportErrorUnknown
	TransportErrorInvalidArgument
	TransportErrorDeadlineExceeded
	TransportErrorNotFound
	TransportErrorAlreadyExists
	TransportErrorPermissionDenied
	TransportErrorResourceExhausted
	TransportErrorFailedPrecondition
	TransportErrorAborted
	TransportErrorOutOfRange
	TransportErrorUnimplemented
	TransportErrorInternal
	TransportErrorUnavailable
	TransportErrorDataLoss
	TransportErrorUnauthenticated
)

var grpcCodesToTransportError = [...]TransportErrorCode{
	grpc.Canceled:           TransportErrorCanceled,
	grpc.Unknown:            TransportErrorUnknown,
	grpc.InvalidArgument:    TransportErrorInvalidArgument,
	grpc.DeadlineExceeded:   TransportErrorDeadlineExceeded,
	grpc.NotFound:           TransportErrorNotFound,
	grpc.AlreadyExists:      TransportErrorAlreadyExists,
	grpc.PermissionDenied:   TransportErrorPermissionDenied,
	grpc.ResourceExhausted:  TransportErrorResourceExhausted,
	grpc.FailedPrecondition: TransportErrorFailedPrecondition,
	grpc.Aborted:            TransportErrorAborted,
	grpc.OutOfRange:         TransportErrorOutOfRange,
	grpc.Unimplemented:      TransportErrorUnimplemented,
	grpc.Internal:           TransportErrorInternal,
	grpc.Unavailable:        TransportErrorUnavailable,
	grpc.DataLoss:           TransportErrorDataLoss,
	grpc.Unauthenticated:    TransportErrorUnauthenticated,
}

func transportErrorCode(c grpc.Code) TransportErrorCode {
	if int(c) < len(grpcCodesToTransportError) {
		return grpcCodesToTransportError[c]
	}
	return TransportErrorUnknownCode
}

var transportErrorToString = [...]string{
	TransportErrorCanceled:           "canceled",
	TransportErrorUnknown:            "unknown",
	TransportErrorInvalidArgument:    "invalid argument",
	TransportErrorDeadlineExceeded:   "deadline exceeded",
	TransportErrorNotFound:           "not found",
	TransportErrorAlreadyExists:      "already exists",
	TransportErrorPermissionDenied:   "permission denied",
	TransportErrorResourceExhausted:  "resource exhausted",
	TransportErrorFailedPrecondition: "failed precondition",
	TransportErrorAborted:            "aborted",
	TransportErrorOutOfRange:         "out of range",
	TransportErrorUnimplemented:      "unimplemented",
	TransportErrorInternal:           "internal",
	TransportErrorUnavailable:        "unavailable",
	TransportErrorDataLoss:           "data loss",
	TransportErrorUnauthenticated:    "unauthenticated",
}

func transportErrorString(t TransportErrorCode) string {
	if int(t) < len(transportErrorToString) {
		return transportErrorToString[t]
	}
	return "unknown code"
}

// IsTransportError reports whether err is TransportError with given code as
// the Reason.
func IsTransportError(err error, code TransportErrorCode) bool {
	if err == nil {
		return false
	}
	var t *TransportError
	if !errors.As(err, &t) {
		return false
	}
	return t.Reason == code
}

func MapGRPCError(err error) error {
	if err == nil {
		return nil
	}
	var t *TransportError
	if errors.As(err, &t) {
		return t
	}
	if s, ok := status.FromError(err); ok {
		return &TransportError{
			Reason:  transportErrorCode(s.Code()),
			message: s.Message(),
			err:     err,
		}
	}
	return err
}

func MustPessimizeEndpoint(err error) bool {
	var t *TransportError
	switch {
	case err == nil:
		return false
	case errors.As(err, &t):
		switch t.Reason {
		case
			TransportErrorResourceExhausted,
			TransportErrorOutOfRange:
			return false
		default:
			return true
		}
	default:
		return false
	}
}
