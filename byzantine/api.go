package byzantine

import (
	"fmt"
)

// ByzantineAPI provides RPC methods
type ByzAPI struct {
	service *ByzantineService
}

// NewByzantineAPI creates a new API instance
func NewByzantineAPI(service *ByzantineService) *ByzAPI {
	return &ByzAPI{service: service}
}

// ConfigureAttack configures a new attack via RPC
func (api *ByzAPI) ConfigureAttack(params AttackParams) (string, error) {
	if api.service.attackManager == nil {
		return "", fmt.Errorf("attack manager not initialized")
	}

	return api.service.attackManager.ConfigureAttack(params)
}

// ListAttacks returns all configured attacks
func (api *ByzAPI) ListAttacks() ([]AttackInfo, error) {
	if api.service.attackManager == nil {
		return nil, fmt.Errorf("attack manager not initialized")
	}

	return api.service.attackManager.ListAttacks(), nil
}

// StopAttack stops a specific attack
func (api *ByzAPI) StopAttack(attackID string) error {
	if api.service.attackManager == nil {
		return fmt.Errorf("attack manager not initialized")
	}

	return api.service.attackManager.StopAttack(attackID)
}
