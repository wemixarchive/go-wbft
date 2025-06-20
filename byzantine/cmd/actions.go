package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/urfave/cli/v2"
)

func byzantineList(ctx *cli.Context) error {
	client, err := rpc.Dial(ctx.String("endpoint"))
	if err != nil {
		return err
	}
	defer client.Close()

	var result interface{}
	if err := client.Call(&result, "byzantine_byzantineTests"); err != nil {
		return err
	}

	// Format output nicely
	return formatByzantineTests(result)
}

func byzantineStop(ctx *cli.Context) error {
	if ctx.NArg() != 1 {
		return fmt.Errorf("expected 1 argument (comma-separated UIDs)")
	}

	uidStrs := strings.Split(ctx.Args().Get(0), ",")
	uids := make([]uint64, len(uidStrs))
	for i, str := range uidStrs {
		uid, err := strconv.ParseUint(strings.TrimSpace(str), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid UID: %s", str)
		}
		uids[i] = uid
	}

	client, err := rpc.Dial(ctx.String("endpoint"))
	if err != nil {
		return err
	}
	defer client.Close()

	var result interface{}
	if err := client.Call(&result, "byzantine_stopByzantineTests", uids); err != nil {
		return err
	}

	fmt.Printf("Stopped tests: %v\n", result)
	return nil
}

func byzantineSilent(ctx *cli.Context) error {
	params := map[string]interface{}{
		"sequence":  ctx.Uint64("sequence"),
		"round":     ctx.Uint64("round"),
		"code":      ctx.Uint64("code"),
		"direction": ctx.Uint64("direction"),
		"targets":   parseTargets(ctx.StringSlice("targets")),
	}

	return callByzantineAPI(ctx, "byzantine_silentMessage", params)
}

func byzantineTamper(ctx *cli.Context) error {
	tamperFields := parseTamperFields(ctx.StringSlice("fields"))

	params := map[string]interface{}{
		"sequence":         ctx.Uint64("sequence"),
		"round":            ctx.Uint64("round"),
		"code":             ctx.String("code"),
		"tamperFields":     tamperFields,
		"withValidMessage": ctx.Bool("with-valid"),
		"delay":            ctx.Uint64("delay"),
		"targets":          parseTargets(ctx.StringSlice("targets")),
	}

	return callByzantineAPI(ctx, "byzantine_sendTamperedMessage", params)
}

func byzantineFake(ctx *cli.Context) error {
	params := map[string]interface{}{
		"sequence": ctx.Uint64("sequence"),
		"round":    ctx.Uint64("round"),
		"code":     ctx.String("code"),
		"targets":  parseTargets(ctx.StringSlice("targets")),
	}

	if msg := ctx.String("message"); msg != "" {
		params["fakeMessage"] = common.FromHex(msg)
	}

	return callByzantineAPI(ctx, "byzantine_sendFakeMessage", params)
}

func byzantineOmit(ctx *cli.Context) error {
	params := map[string]interface{}{
		"sequence": ctx.Uint64("sequence"),
		"round":    ctx.Uint64("round"),
		"code":     ctx.String("code"),
		"cmd":      ctx.Uint64("cmd"),
		"cnt":      ctx.Uint64("count"),
		"targets":  parseTargets(ctx.StringSlice("targets")),
	}

	return callByzantineAPI(ctx, "byzantine_sendOmitMessage", params)
}

func byzantineRoleSpoof(ctx *cli.Context) error {
	params := map[string]interface{}{
		"sequence": ctx.Uint64("sequence"),
		"round":    ctx.Uint64("round"),
		"code":     ctx.String("code"),
		"targets":  parseTargets(ctx.StringSlice("targets")),
	}

	if msg := ctx.String("message"); msg != "" {
		params["fakeMessage"] = common.FromHex(msg)
	}

	return callByzantineAPI(ctx, "byzantine_sendRoleSpoofedMessage", params)
}

func byzantineReplay(ctx *cli.Context) error {
	params := map[string]interface{}{
		"ori_sequence":    ctx.Uint64("orig-sequence"),
		"ori_round":       ctx.Uint64("orig-round"),
		"sequence":        ctx.Uint64("sequence"),
		"round":           ctx.Uint64("round"),
		"useOriginalView": ctx.Bool("use-original-view"),
		"code":            ctx.String("code"),
		"targets":         parseTargets(ctx.StringSlice("targets")),
	}

	return callByzantineAPI(ctx, "byzantine_sendReplayMessage", params)
}

// Helper functions

func parseTargets(targets []string) []common.Address {
	addrs := make([]common.Address, 0, len(targets))
	for _, target := range targets {
		if addr := common.HexToAddress(target); addr != (common.Address{}) {
			addrs = append(addrs, addr)
		}
	}
	return addrs
}

func parseTamperFields(fields []string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(fields))
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) == 2 {
			result = append(result, map[string]interface{}{
				"target": parts[0],
				"value":  parts[1],
			})
		}
	}
	return result
}

func callByzantineAPI(ctx *cli.Context, method string, params map[string]interface{}) error {
	client, err := rpc.Dial(ctx.String("endpoint"))
	if err != nil {
		return err
	}
	defer client.Close()

	var result interface{}
	if err := client.Call(&result, method, params); err != nil {
		return err
	}

	fmt.Printf("Result: %v\n", result)
	return nil
}

// formatByzantineTests formats the Byzantine tests output nicely
func formatByzantineTests(result interface{}) error {
	if result == nil {
		fmt.Println("No Byzantine tests registered.")
		return nil
	}

	// Try to parse as JSON first for better formatting
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err == nil {
		// Check if it's an empty array
		var tests []map[string]interface{}
		if json.Unmarshal(jsonBytes, &tests) == nil {
			if len(tests) == 0 {
				fmt.Println("No Byzantine tests registered.")
				return nil
			}

			// Format as table
			fmt.Printf("Registered Byzantine Tests (%d):\n", len(tests))
			fmt.Printf("%-5s %-20s %-8s %-10s %-8s %-10s\n",
				"UID", "Name", "Type", "Sequence", "Round", "Status")
			fmt.Printf("%-5s %-20s %-8s %-10s %-8s %-10s\n",
				"---", "----", "----", "--------", "-----", "------")

			for _, test := range tests {
				uid := getStringValue(test, "uid")
				name := getStringValue(test, "name")
				testType := getStringValue(test, "type")
				sequence := getStringValue(test, "sequence")
				round := getStringValue(test, "round")
				status := getStringValue(test, "status")

				// Truncate long names
				if len(name) > 20 {
					name = name[:17] + "..."
				}

				fmt.Printf("%-5s %-20s %-8s %-10s %-8s %-10s\n",
					uid, name, testType, sequence, round, status)
			}
			return nil
		}
	}

	// Fallback to JSON output if table formatting fails
	if jsonBytes != nil {
		fmt.Printf("Registered Byzantine Tests:\n%s\n", string(jsonBytes))
	} else {
		fmt.Printf("Registered Byzantine Tests:\n%v\n", result)
	}
	return nil
}

// getStringValue safely gets a string value from a map
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case float64:
			return strconv.FormatFloat(v, 'f', 0, 64)
		case int:
			return strconv.Itoa(v)
		case int64:
			return strconv.FormatInt(v, 10)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}
