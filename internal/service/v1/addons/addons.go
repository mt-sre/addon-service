package addons

import (
	v1 "github.com/mt-sre/addon-service/api/v1"
)

type AddonsService struct {
	v1.UnimplementedAddonServiceServer
}

func NewAddonService() *AddonsService {
	return &AddonsService{}
}
