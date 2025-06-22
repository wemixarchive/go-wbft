package cmd

import "github.com/urfave/cli/v2"

var (
	byzListCommand = &cli.Command{
		Name:   "list",
		Usage:  "List all registered Byzantine tests",
		Action: byzantineList,
		Flags: []cli.Flag{
			RpcEndpoint,
		},
	}

	byzStopCommand = &cli.Command{
		Name:      "stop",
		Usage:     "Stop Byzantine tests by UIDs",
		ArgsUsage: "<uid1,uid2,...>",
		Action:    byzantineStop,
		Flags: []cli.Flag{
			RpcEndpoint,
		},
	}
	byzSilentCommand = &cli.Command{
		Name:   "silent",
		Usage:  "Configure silent message attack",
		Action: byzantineSilent,
		Flags: append(byzantineAttackFlags(), []cli.Flag{
			MessageCodeFlag,
			InOutDirectionFlag,
		}...),
	}
	byzTamperCommand = &cli.Command{
		Name:   "tamper",
		Usage:  "Configure tampered message attack",
		Action: byzantineTamper,
		Flags: append(byzantineAttackFlags(), []cli.Flag{
			MessageCodeFlag,
			TamperFieldFlag,
			TamperWithValidFlag,
			OutputDelayFlag,
		}...),
	}

	byzFakeCommand = &cli.Command{
		Name:   "fake",
		Usage:  "Configure fake message attack",
		Action: byzantineFake,
		Flags: append(byzantineAttackFlags(), []cli.Flag{
			MessageCodeFlag,
			FakeMessageFlag,
		}...),
	}
	byzOmitCommand = &cli.Command{
		Name:   "omit",
		Usage:  "Configure omit message attack",
		Action: byzantineOmit,
		Flags: append(byzantineAttackFlags(), []cli.Flag{
			MessageCodeFlag,
			OmitMessageCmdFlag,
			OmitMessageCountFlag,
		}...),
	}

	byzRoleSpoofCommand = &cli.Command{
		Name:   "rolespoof",
		Usage:  "Configure role spoofing attack",
		Action: byzantineRoleSpoof,
		Flags: append(byzantineAttackFlags(), []cli.Flag{
			MessageCodeFlag,
			FakeMessageFlag,
		}...),
	}

	byzReplayCommand = &cli.Command{
		Name:   "replay",
		Usage:  "Configure replay attack",
		Action: byzantineReplay,
		Flags: append(byzantineAttackFlags(), []cli.Flag{
			OriginSeqFlag,
			OriginRoundFlag,
			UseOriginViewFlag,
			MessageCodeFlag,
		}...),
	}
)

// ByzantineCommand returns the Byzantine main command
func ByzantineCommand() *cli.Command {
	return &cli.Command{
		Name:     "byzantine",
		Usage:    "Byzantine fault simulation commands",
		Category: BYZCategory,
		Subcommands: []*cli.Command{
			byzListCommand,
			byzStopCommand,
			byzSilentCommand,
			byzTamperCommand,
			byzFakeCommand,
			byzOmitCommand,
			byzRoleSpoofCommand,
			byzReplayCommand,
		},
		Description: `
The byzantine command suite provides tools for simulating Byzantine faults in the QBFT consensus.
These commands allow you to test the resilience of the consensus algorithm against various attack scenarios.`,
	}
}
