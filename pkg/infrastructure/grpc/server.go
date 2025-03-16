package grpc

import (
	"context"
	"dice-game/pkg/domain/interfaces"
	"dice-game/pkg/usecase"
	pb "dice-game/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"net"
)

type Server struct {
	address     string
	logger      interfaces.Logger
	server      *grpc.Server
	gameUseCase usecase.GameUseCaseInterface
}

func NewServer(address string, logger interfaces.Logger, gameUseCase usecase.GameUseCaseInterface) *Server {
	return &Server{
		address:     address,
		logger:      logger.With().Str("component", "grpc_server").Logger(),
		gameUseCase: gameUseCase,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to listen")
		return err
	}

	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(s.panicRecoveryInterceptor()),
	}
	s.server = grpc.NewServer(opts...)

	diceGameService := NewDiceGameService(s.gameUseCase, s.logger)
	pb.RegisterDiceGameServiceServer(s.server, diceGameService)

	reflection.Register(s.server)

	s.logger.Info().Str("address", s.address).Msg("gRPC server starting")

	if err := s.server.Serve(listener); err != nil {
		s.logger.Error().Err(err).Msg("Failed to serve")
		return err
	}

	return nil
}

func (s *Server) panicRecoveryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error().Interface("panic", r).Str("method", info.FullMethod).Msg("Recovered from panic")
				err = status.Errorf(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}

func (s *Server) Stop() {
	if s.server != nil {
		s.logger.Info().Msg("Gracefully stopping gRPC server")
		s.server.GracefulStop()
		s.logger.Info().Msg("gRPC server stopped")
	}
}
