package config

import "nodes-indexer/modules/config/dto"

type ConfigService struct {
	*dto.Config
}

func NewConfigService(config *dto.Config) ConfigService {
	return ConfigService{Config: config}
}