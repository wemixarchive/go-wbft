package api

import (
	"github.com/ethereum/go-ethereum/byzantine/types"
	"github.com/ethereum/go-ethereum/log"
)

// PublicByzantineAPI provides the public RPC interface for Byzantine module
type PublicByzantineAPI struct {
	handler *Handler
}

// NewPublicByzantineAPI creates a new RPC API instance
func NewPublicByzantineAPI(service types.ByzantineService) *PublicByzantineAPI {
	return &PublicByzantineAPI{
		handler: NewHandler(service),
	}
}

// ByzantineTests returns all registered Byzantine tests
// This matches the web3ext.js getter definition
func (api *PublicByzantineAPI) ByzantineTests() ([]types.AttackConfig, error) {
	return api.handler.GetByzantineTests()
}

// StopByzantineTests stops Byzantine tests by UIDs
func (api *PublicByzantineAPI) StopByzantineTests(uids []string) error {
	log.Info("[Byzantine API] StopByzantineTests called", "uids", uids)
	return api.handler.StopByzantineTests(uids)
}

// SilentMessage configures a silent message attack
func (api *PublicByzantineAPI) SilentMessage(params map[string]interface{}) error {
	log.Info("[Byzantine API] SilentMessage called")

	// Convert raw params to typed struct
	typedParams, err := ConvertToSilentMessageParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterSilentMessage(typedParams)
}

// SendTamperedMessage configures a tampered message attack
func (api *PublicByzantineAPI) SendTamperedMessage(params map[string]interface{}) error {
	log.Info("[Byzantine API] SendTamperedMessage called")

	typedParams, err := ConvertToTamperedMessageParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterTamperedMessage(typedParams)
}

// SendFakeMessage configures a fake message attack
func (api *PublicByzantineAPI) SendFakeMessage(params map[string]interface{}) error {
	log.Info("[Byzantine API] SendFakeMessage called")

	typedParams, err := ConvertToFakeMessageParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterFakeMessage(typedParams)
}

// SendOmitMessage configures an omit message attack
func (api *PublicByzantineAPI) SendOmitMessage(params map[string]interface{}) error {
	log.Info("[Byzantine API] SendOmitMessage called")

	typedParams, err := ConvertToOmitMessageParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterOmitMessage(typedParams)
}

// SendRoleSpoofedMessage configures a role spoofing attack
func (api *PublicByzantineAPI) SendRoleSpoofedMessage(params map[string]interface{}) error {
	log.Info("[Byzantine API] SendRoleSpoofedMessage called")

	typedParams, err := ConvertToRoleSpoofParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterRoleSpoofedMessage(typedParams)
}

// SendReplayMessage configures a replay attack
func (api *PublicByzantineAPI) SendReplayMessage(params map[string]interface{}) error {
	log.Info("[Byzantine API] SendReplayMessage called")

	typedParams, err := ConvertToReplayMessageParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterReplayMessage(typedParams)
}

// UpgradeGovContract upgrades governance contract
func (api *PublicByzantineAPI) UpgradeGovContract() error {
	log.Info("[Byzantine API] UpgradeGovContract called")
	return api.handler.UpgradeGovContract()
}
