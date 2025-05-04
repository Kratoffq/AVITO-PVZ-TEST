package grpc

import (
	"context"

	"github.com/avito/pvz/api/proto"
	servicePVZ "github.com/avito/pvz/internal/service/pvz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PVZHandler реализует gRPC-интерфейс для работы с ПВЗ
type PVZHandler struct {
	proto.UnimplementedPVZServiceServer
	pvzService *servicePVZ.Service
}

// NewPVZHandler создает новый экземпляр PVZHandler
func NewPVZHandler(pvzService *servicePVZ.Service) *PVZHandler {
	return &PVZHandler{
		pvzService: pvzService,
	}
}

// GetAllPVZ возвращает список всех ПВЗ
func (h *PVZHandler) GetAllPVZ(ctx context.Context, req *proto.GetAllPVZRequest) (*proto.GetAllPVZResponse, error) {
	pvzs, err := h.pvzService.GetAll(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get PVZs")
	}

	response := &proto.GetAllPVZResponse{
		Pvzs: make([]*proto.PVZ, len(pvzs)),
	}

	for i, p := range pvzs {
		response.Pvzs[i] = &proto.PVZ{
			Id:               p.ID.String(),
			City:             p.City,
			RegistrationDate: timestamppb.New(p.CreatedAt),
		}
	}

	return response, nil
}
