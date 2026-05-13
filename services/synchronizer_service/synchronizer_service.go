package synchronizerservice

import (
	jobs_service "github.com/florin-rada/grm-back/services/jobs_service"
	runners_service "github.com/florin-rada/grm-back/services/runners_service"
)

type SynchronizerService struct {
	rs *runners_service.RunnersService
	js *jobs_service.JobsService
}

func NewSynchronizerService(rs *runners_service.RunnersService, js *jobs_service.JobsService) *SynchronizerService {
	return &SynchronizerService{rs: rs, js: js}
}
