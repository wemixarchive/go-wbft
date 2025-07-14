// Package attacks implements various Byzantine attack types.
// This file ensures all attack implementations are included when the package is imported.
package attacks

// This package contains the following attack implementations:
// - SilentMessageAttack: Drops messages silently
// - TamperedMessageAttack: Sends tampered messages
// - FakeMessageAttack: Sends fake messages
// - OmitMessageAttack: Omits parts of messages
// - RoleSpoofAttack: Spoofs validator roles
// - ReplayAttack: Replays old messages (currently disabled due to storage removal)
// - Store message: Store messages and reuse them to perform Byzantine attacks

// Each attack registers itself via init() functions in their respective files.
