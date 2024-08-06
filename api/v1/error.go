package log_v1

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type ErrOffsetOutOfRange struct {
	Offset uint64
}

func (e ErrOffsetOutOfRange) GRPCStatus() *status.Status {
	status := status.New(codes.OutOfRange, fmt.Sprintf("offset %d not found", e.Offset))
	msg := fmt.Sprintf("The requested offset is outside the log's range. The log has a maximum offset of %d.", e.Offset)
	d := &errdetails.LocalizedMessage{
		Locale:  "en-US",
		Message: msg,
	}
	std, err := status.WithDetails(d)
	if err != nil {
		return status
	}
	return std
}

func (e ErrOffsetOutOfRange) Error() string {
	return e.GRPCStatus().Err().Error()
}
