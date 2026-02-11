// internal/services/security.go
package services

import "context"

type SecurityService struct {
	ctx context.Context
}

func NewSecurityService(queries any) *SecurityService {
	return &SecurityService{}
}
