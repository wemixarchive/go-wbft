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
	log.Info("BYZ: StopByzantineTests called", "uids", uids)
	return api.handler.StopByzantineTests(uids)
}

// SetMessagePolicy controls original message sending behavior per sequence/round,
// allowing message drops or delays to simulate Byzantine scenarios.
func (api *PublicByzantineAPI) SetMessagePolicy(params map[string]interface{}) error {
	log.Info("BYZ: SetMessagePolicy called")

	typedParams, err := ConvertToSetMessagePolicyParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterSetMessagePolicy(typedParams)
}

// SendTamperedMessage configures a tampered message attack
func (api *PublicByzantineAPI) SendTamperedMessage(params map[string]interface{}) error {
	log.Info("BYZ: SendTamperedMessage called")

	typedParams, err := ConvertToTamperedMessageParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterTamperedMessage(typedParams)
}

// SendFakeMessage configures a fake message attack
func (api *PublicByzantineAPI) SendFakeMessage(params map[string]interface{}) error {
	log.Info("BYZ: SendFakeMessage called")

	typedParams, err := ConvertToFakeMessageParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterFakeMessage(typedParams)
}

// SendOmitMessage configures an omit message attack
func (api *PublicByzantineAPI) SendOmitMessage(params map[string]interface{}) error {
	log.Info("BYZ: SendOmitMessage called")

	typedParams, err := ConvertToOmitMessageParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterOmitMessage(typedParams)
}

// SendRoleSpoofedMessage configures a role spoofing attack
func (api *PublicByzantineAPI) SendRoleSpoofedMessage(params map[string]interface{}) error {
	log.Info("BYZ: SendRoleSpoofedMessage called")

	typedParams, err := ConvertToRoleSpoofParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterRoleSpoofedMessage(typedParams)
}

// SendReplayMessage configures a replay attack
func (api *PublicByzantineAPI) SendReplayMessage(params map[string]interface{}) error {
	log.Info("BYZ: SendReplayMessage called")

	typedParams, err := ConvertToReplayMessageParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterReplayMessage(typedParams)
}

// StoreMessage stores messages that can later be reused to perform Byzantine attacks.
func (api *PublicByzantineAPI) StoreMessage(params map[string]interface{}) error {
	log.Info("BYZ: StoreMessage called")

	typedParams, err := ConvertToStoreMessageParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterStoreMessage(typedParams)
}

// SendDosMessage configures a DoS message flooding attack
func (api *PublicByzantineAPI) SendDosMessage(params map[string]interface{}) error {
	log.Info("BYZ: SendDosMessage called")

	typedParams, err := ConvertToDosMessageParams(params)
	if err != nil {
		return err
	}

	return api.handler.RegisterDosMessage(typedParams)
}

// UpgradeGovContract upgrades governance contract
func (api *PublicByzantineAPI) UpgradeGovContract() error {
	log.Info("BYZ: UpgradeGovContract called")
	return api.handler.UpgradeGovContract()
}

// RegisterAttacks registers multiple attacks at once (like loadAttacksFromConfig)
func (api *PublicByzantineAPI) RegisterAttacks(params types.RegisterAttacksParams) (*types.BatchRegisterResponse, error) {
	log.Info("BYZ: RegisterAttacks called", "count", len(params.Attacks))
	return api.handler.RegisterAttacks(params)
}

// GetActiveAttacks returns all active attacks
func (api *PublicByzantineAPI) GetActiveAttacks() ([]types.AttackStatusResponse, error) {
	log.Info("BYZ: GetActiveAttacks called")
	return api.handler.GetActiveAttacks()
}

// GetAttackStatus returns the status of a specific attack
func (api *PublicByzantineAPI) GetAttackStatus(uid string) (*types.AttackStatusResponse, error) {
	log.Info("BYZ: GetAttackStatus called", "uid", uid)
	return api.handler.GetAttackStatus(uid)
}

// GetAttackMetrics returns metrics about attacks
func (api *PublicByzantineAPI) GetAttackMetrics() (*types.Metrics, error) {
	log.Info("BYZ: GetAttackMetrics called")
	return api.handler.GetAttackMetrics()
}
