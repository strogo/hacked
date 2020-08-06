package archive

import (
	"fmt"

	"github.com/inkyblackness/hacked/ss1/content/interpreters"
	"github.com/inkyblackness/hacked/ss1/content/text"
)

// GameStateSize specifies the byte count of a serialized GameState.
const GameStateSize = 0x054D
const (
	stateHackerNameSize  = 20
	inventoryWeaponSlots = 7
	grenadeTypeCount     = 7

	engineTicksPerSecond = 280
	secondsPerMinute     = 60
	engineTicksPerMinute = secondsPerMinute * engineTicksPerSecond
	cyberspaceMinTime    = 90 * engineTicksPerSecond
	cyberspaceMaxTime    = 30 * secondsPerMinute * engineTicksPerSecond

	messageStatusReceived = 0x80
	messageStatusRead     = 0x40

	// BooleanVarCount is the number of avaiable boolean variables.
	BooleanVarCount = 512
	IntegerVarCount = 64
)

type GameState struct {
	*interpreters.Instance
}

var gameStateDesc = interpreters.New().
	With("Difficulty: Combat", 0x0015, 1).As(interpreters.RangedValue(0, 3)).
	With("Difficulty: Mission", 0x0016, 1).As(interpreters.RangedValue(0, 3)).
	With("Difficulty: Puzzle", 0x0017, 1).As(interpreters.RangedValue(0, 3)).
	With("Difficulty: Cyber", 0x0018, 1).As(interpreters.RangedValue(0, 3)).
	With("Game time", 0x001C, 4).As(interpreters.SpecialValue("Internal")).
	With("Current Level", 0x0039, 1).As(interpreters.RangedValue(0, 15)).
	With("Health", 0x009C, 1).As(interpreters.FormattedRangedValue(0, 255,
	func(value int) string {
		return fmt.Sprintf("%.2f%%%%", float64(value*100)/255.0)
	})).
	With("Cyberspace integrity", 0x009D, 1).As(interpreters.FormattedRangedValue(0, 255,
	func(value int) string {
		return fmt.Sprintf("%.2f%%%%", float64(value*100)/255.0)
	})).
	With("Health regeneration per minute", 0x009E, 2).As(interpreters.RangedValue(0, 1000)).
	With("Power", 0x00AC, 1).As(interpreters.FormattedRangedValue(0, 255,
	func(value int) string {
		return fmt.Sprintf("%.2f%%%%", float64(value*100)/255.0)
	})).
	With("Out-of-power flag", 0x00AF, 1).As(interpreters.EnumValue(
	map[uint32]string{
		0: "Powered",
		1: "Out-of-power",
	})).
	With("Cyberspace base time", 0x00B2, 4).As(interpreters.FormattedRangedValue(cyberspaceMinTime, cyberspaceMaxTime,
	func(value int) string {
		minutes := value / engineTicksPerMinute
		return fmt.Sprintf("%2dm %2.2fs", minutes, float64(value-(minutes*engineTicksPerMinute))/(engineTicksPerSecond))
	})).
	With("Fatigue regeneration", 0x0181, 2).As(interpreters.RangedValue(0, 400)).
	With("Fatigue regeneration base", 0x0183, 2).As(interpreters.RangedValue(0, 400)).
	With("Fatigue regeneration max", 0x0185, 2).As(interpreters.RangedValue(0, 400)).
	With("Accuracy", 0x0187, 1).As(interpreters.RangedValue(0, 255)).
	With("Shield absorb rate", 0x0188, 1).As(interpreters.RangedValue(0, 255)).
	With("Hardware: Infrared", 0x2E9, 1).As(interpreters.RangedValue(0, 4)).
	With("Hardware: Targeting", 0x2EA, 1).As(interpreters.RangedValue(0, 4)).
	With("Hardware: Sensaround", 0x2EB, 1).As(interpreters.RangedValue(0, 4)).
	With("Hardware: Aim enhancement", 0x2EC, 1).As(interpreters.RangedValue(0, 4)).
	With("Hardware: Bioscan", 0x2EE, 1).As(interpreters.RangedValue(0, 4)).
	With("Hardware: Navigation unit", 0x2EF, 1).As(interpreters.RangedValue(0, 4)).
	With("Hardware: Shield", 0x2F0, 1).As(interpreters.RangedValue(0, 4)).
	With("Hardware: Data reader", 0x2F1, 1).As(interpreters.RangedValue(0, 4)).
	With("Hardware: Lantern", 0x2F2, 1).As(interpreters.RangedValue(0, 4)).
	With("Hardware: Enviro suit", 0x2F4, 1).As(interpreters.RangedValue(0, 4)).
	With("Hardware: Booster", 0x2F5, 1).As(interpreters.RangedValue(0, 4)).
	With("Hardware: Jump jet", 0x2F6, 1).As(interpreters.RangedValue(0, 4)).
	With("Hardware: System status", 0x2F7, 1).As(interpreters.RangedValue(0, 4))

// NewGameState() returns a GameState instance for given raw data.
func NewGameState(raw []byte) GameState {
	return GameState{Instance: gameStateDesc.For(raw)}
}

// IsSavegame returns true if the state describes one during a running game.
func (state *GameState) IsSavegame() bool {
	return state.Get("Game time") > 0
}

// HackerName returns the name interpreted with given codepage.
func (state *GameState) HackerName(cp text.Codepage) string {
	return cp.Decode(state.Raw()[:stateHackerNameSize])
}

// SetHackerName stores the given name using given codepage, up to the internal available amount of bytes.
func (state *GameState) SetHackerName(name string, cp text.Codepage) {
	raw := state.Raw()
	var zeroName [stateHackerNameSize]byte
	copy(raw[:stateHackerNameSize], zeroName[:])
	copy(raw[:stateHackerNameSize], cp.Encode(name))
	raw[stateHackerNameSize-1] = 0
}

// BooleanVar returns the state of the boolean variable at given index. Unsupported indices return 0.
func (state GameState) BooleanVar(index int) bool {
	if (index < 0) || (index >= BooleanVarCount) {
		return false
	}
	return state.Raw()[0x00B6+(index/8)]&(0x01<<(uint(index%8))) != 0
}

// SetBooleanVar sets the state of the boolean variable at given index. Unsupported indices are ignored.
func (state *GameState) SetBooleanVar(index int, set bool) {
	if (index < 0) || (index >= BooleanVarCount) {
		return
	}
	byteIndex := 0x00B6 + (index / 8)
	bitMask := byte(0x01 << uint(index%8))
	temp := state.Raw()[byteIndex]
	temp &= ^bitMask
	if set {
		temp |= bitMask
	}
	state.Raw()[byteIndex] = temp
}

// IntegerVar returns the value of the integer variable at given index. Unsupported indices return 0.
func (state *GameState) IntegerVar(index int) int16 {
	if (index < 0) || (index >= IntegerVarCount) {
		return 0
	}
	startOffset := 0x00F6 + (2 * index)
	val := uint16(state.Raw()[startOffset+1])<<8 | uint16(state.Raw()[startOffset+0])
	return int16(val)
}

// SetIntegerVar sets the value of the integer variable at given index. Unsupported indices are ignored.
func (state *GameState) SetIntegerVar(index int, value int16) {
	if (index < 0) || (index >= IntegerVarCount) {
		return
	}
	startOffset := 0x00F6 + (2 * index)
	state.Raw()[startOffset+0] = byte((value >> 0) & 0xFF)
	state.Raw()[startOffset+1] = byte((value >> 8) & 0xFF)
}

// DefaultGameStateData returns the state block initialized as if the engine started a new default game.
func DefaultGameStateData() []byte {
	data := ZeroGameStateData()
	state := NewGameState(data)

	state.Set("Health", 212)
	state.Set("Cyberspace integrity", 255)
	state.Set("Health regeneration per minute", 0)
	state.Set("Power", 255)
	state.Set("Cyberspace base time", 30*60*engineTicksPerSecond)
	state.Set("Shield absorb rate", 0)
	state.Set("Fatigue regeneration", 0)
	state.Set("Fatigue regeneration base", 100)
	state.Set("Fatigue regeneration max", 400)
	state.Set("Accuracy", 100)

	for i := 0; i < inventoryWeaponSlots; i++ {
		data[0x046B+(5*i)+0] = 0xFF
	}
	for i := 0; i < grenadeTypeCount; i++ {
		data[0x04DF+(i*2)+1] = 70
	}

	setInitialCitadelHackerState(&state)
	setInitialCitadelVariables(&state)

	return data
}

// ZeroGameStateData returns the state block reset for default engine behaviour.
func ZeroGameStateData() []byte {
	return make([]byte, GameStateSize)
}

func setInitialCitadelHackerState(state *GameState) {
	data := state.Raw()

	state.Set("Current Level", 1)

	// location: in the neurosurgery chamber, looking East
	data[0x051F+0+1] = 0x80
	data[0x051F+0+2] = 0x1E
	data[0x051F+4+1] = 0x80
	data[0x051F+4+2] = 0x16
	data[0x051F+8+1] = 0xBD
	data[0x051F+8+2] = 0x01 // TODO: this height is level shift dependent!
	// rotation:
	data[0x051F+12+0] = 0x3E
	data[0x051F+12+1] = 0x24
	data[0x051F+12+2] = 0x03

	// set first message
	data[0x0337+26] = messageStatusReceived // Rebecca Lansing's first message
	data[0x04F9+9] = 0xFF                   // HUD active email -- set for similarity, has no effect.
}

func setInitialCitadelVariables(state *GameState) {
	// The following set is taken from player.c
	initialBooleanVariables := []int{
		0x001, 0x002, 0x003, 0x010, 0x012, 0x015, 0x016, 0x017, 0x018, 0x019, 0x01A,
		0x020, 0x021, 0x024, 0x025,
		0x04B, 0x04C, 0x04D, 0x04E, 0x04F,
		0x050, 0x051, 0x052, 0x053, 0x054, 0x055, 0x056, 0x057, 0x058, 0x059, 0x05A, 0x05B, 0x05C, 0x05D, 0x05E, 0x05F,
		0x070, 0x071, 0x072, 0x073, 0x074, 0x075, 0x076, 0x077, 0x078, 0x079, 0x07A, 0x07B, 0x07C, 0x07D, 0x07E, 0x07F,
		0x0A0, 0x0A1, 0x0A2, 0x0A3, 0x0A4, 0x0A5, 0x0A6, 0x0A7, 0x0A8, 0x0A9,
		0x0C0, 0x0C1, 0x0C2, 0x0C3, 0x0C4, 0x0C5, 0x0C6, 0x0C7, 0x0C8, 0x0C9, 0x0CA, 0x0CB, 0x0CC, 0x0CD, 0x0CE, 0x0CF,
		0x0E1, 0x0E3, 0x0E5, 0x0E7, 0x0E9, 0x0EB, 0x0ED, 0x0EF,
		0x0F1, 0x0F3, 0x0F5, 0x0F7, 0x0F9, 0x0FB, 0x0FD, 0x0FF,
		0x101, 0x103, 0x105, 0x107, 0x109, 0x10B, 0x10D, 0x10F,
		0x111, 0x113, 0x115, 0x117, 0x119, 0x11B, 0x11D, 0x11F,
		0x121, 0x123, 0x125, 0x127, 0x129, 0x12B,
	}
	initialIntegerVariables := map[int]int16{
		0x03: 2,     // engine state
		0x0C: 3,     // number of available groves
		0x33: 0x100, // joystick sensitivity
	}
	for i := 0; i < BooleanVarCount; i++ {
		state.SetBooleanVar(i, false)
	}
	for _, index := range initialBooleanVariables {
		state.SetBooleanVar(index, true)
	}
	for i := 0; i < IntegerVarCount; i++ {
		state.SetIntegerVar(i, initialIntegerVariables[i])
	}
}
