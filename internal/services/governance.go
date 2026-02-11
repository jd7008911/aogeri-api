// internal/services/governance.go
package services

import "context"

type GovernanceService struct {
	ctx context.Context
}

func NewGovernanceService(queries any, auth any) *GovernanceService {
	return &GovernanceService{}
}
