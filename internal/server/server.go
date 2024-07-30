package server

import (
	"context"

	api "github.com/MartinMinkov/proglog/api/v1"
	"google.golang.org/grpc"
)

type Config struct {
	CommitLog CommitLog
}

type CommitLog interface {
	Append(record *api.Record) (uint64, error)
	Read(offset uint64) (*api.Record, error)
}

var _ api.LogServer = (*grpcServer)(nil)

type grpcServer struct {
	api.UnimplementedLogServer
	*Config
}

func NewGRPCServer(config *Config) (*grpc.Server, error) {
	grpcServer := grpc.NewServer()
	server, err := newgrpcServer(config)
	if err != nil {
		return nil, err
	}
	api.RegisterLogServer(grpcServer, server)
	return grpcServer, nil
}

func newgrpcServer(config *Config) (*grpcServer, error) {
	server := &grpcServer{
		Config: config,
	}
	return server, nil
}

func (s *grpcServer) Produce(ctx context.Context, req *api.ProduceRequest) (*api.ProduceResponse, error) {
	offset, err := s.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &api.ProduceResponse{
		Offset: offset,
	}, nil
}

func (s *grpcServer) Consume(ctx context.Context, req *api.ConsumeRequest) (*api.ConsumeResponse, error) {
	record, err := s.CommitLog.Read(req.Offset)
	if err != nil {
		return nil, err
	}
	return &api.ConsumeResponse{
		Record: record,
	}, nil
}

func (s *grpcServer) ProduceStream(stream api.Log_ProduceStreamServer) error {
	for {
		request, err := stream.Recv()
		if err != nil {
			return err
		}
		response, err := s.Produce(stream.Context(), request)
		if err != nil {
			return err
		}
		if err := stream.Send(response); err != nil {
			return err
		}
	}
}

func (s *grpcServer) ConsumeStream(request *api.ConsumeRequest, stream api.Log_ConsumeStreamServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			response, err := s.Consume(stream.Context(), request)
			switch err.(type) {
			case nil:
			case api.ErrOffsetOutOfRange:
				continue
			default:
				return err
			}
			if err = stream.Send(response); err != nil {
				return err
			}
			request.Offset++
		}
	}
}
