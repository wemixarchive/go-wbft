package cmd

import "github.com/urfave/cli/v2"

const (
	BYZCategory = "BYZANTINE"
)

var (
	RpcEndpoint = &cli.StringFlag{
		Name:  "endpoint",
		Usage: "RPC endpoint",
		Value: "http://localhost:8545",
	}
	SequenceFlag = &cli.Uint64Flag{
		Name:     "sequence",
		Usage:    "Block sequence number",
		Required: true,
	}
	RoundFlag = &cli.Uint64Flag{
		Name:     "round",
		Usage:    "Round number",
		Required: true,
	}
	MessageCodeFlag = &cli.Uint64Flag{
		Name:     "code",
		Usage:    "Message code bitmap (1=PrePrepare, 2=Prepare, 4=Commit, 8=RoundChange, 16=Propagation)",
		Required: true,
	}
	TargetFlag = &cli.StringSliceFlag{
		Name:  "targets",
		Usage: "Target validator addresses",
	}
	InOutDirectionFlag = &cli.Uint64Flag{
		Name:  "direction",
		Usage: "Direction (1=send, 2=receive, 3=both)",
		Value: 1,
	}
	TamperFieldFlag = &cli.StringSliceFlag{
		Name:     "fields",
		Usage:    "Fields to tamper (format: field=value)",
		Required: true,
	}
	TamperWithValidFlag = &cli.BoolFlag{
		Name:  "with-valid",
		Usage: "Send valid message after tampered one",
		Value: false,
	}
	OutputDelayFlag = &cli.Uint64Flag{
		Name:  "delay",
		Usage: "Delay in milliseconds for valid message",
		Value: 0,
	}
	FakeMessageFlag = &cli.StringFlag{
		Name:  "message",
		Usage: "Fake message content (hex encoded)",
	}
	OmitMessageCmdFlag = &cli.Uint64Flag{
		Name:     "cmd",
		Usage:    "Omit command (varies by message type)",
		Required: true,
	}
	OmitMessageCountFlag = &cli.Uint64Flag{
		Name:  "count",
		Usage: "Number of items to omit (0=all)",
		Value: 0,
	}
	OriginSeqFlag = &cli.Uint64Flag{
		Name:     "orig-sequence",
		Usage:    "Original message sequence number",
		Required: true,
	}
	OriginRoundFlag = &cli.Uint64Flag{
		Name:     "orig-round",
		Usage:    "Original message round number",
		Required: true,
	}
	UseOriginViewFlag = &cli.BoolFlag{
		Name:  "use-original-view",
		Usage: "Use original sequence/round in replayed message",
		Value: false,
	}

	// Byzantine command line flags
	ByzantineEnabledFlag = &cli.BoolFlag{
		Name:     "byzantine",
		Usage:    "Enable Byzantine fault simulation module",
		Value:    true,
		Category: BYZCategory,
	}
	ByzantineConfigFileFlag = &cli.StringFlag{
		Name:     "byzantine.config",
		Usage:    "Path to Byzantine configuration file",
		Category: BYZCategory,
	}
	ByzantineLogLevelFlag = &cli.StringFlag{
		Name:     "byzantine.loglevel",
		Usage:    "Byzantine module log level (trace, debug, info, warn, error, crit)",
		Value:    "info",
		Category: BYZCategory,
	}
)

// byzantineAttackFlags returns common flags for Byzantine attacks
func byzantineAttackFlags() []cli.Flag {
	return []cli.Flag{
		RpcEndpoint,
		SequenceFlag,
		RoundFlag,
		TargetFlag,
	}
}

// ByzantineCommandFlags returns all Byzantine-related flags
func ByzantineCommandFlags() []cli.Flag {
	return []cli.Flag{
		ByzantineEnabledFlag,
		ByzantineConfigFileFlag,
		ByzantineLogLevelFlag,
	}
}
