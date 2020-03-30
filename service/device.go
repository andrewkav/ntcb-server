package service

import (
	"ntcb-server/restmodels"

	"github.com/jinzhu/gorm"
	"github.com/rs/zerolog"
)

type DeviceService struct {
	db     *gorm.DB
	logger zerolog.Logger
}

func (svc *DeviceService) List() ([]restmodels.Device, error) {

	return nil, nil
}
