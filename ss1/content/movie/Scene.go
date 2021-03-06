package movie

import (
	"time"

	"github.com/inkyblackness/hacked/ss1/content/bitmap"
)

// Scene describes a series of frames that share a common palette.
// Ideally, they also share the same framerate.
type Scene struct {
	Palette bitmap.Palette
	Frames  []Frame
}

// Frame describes a bitmap and how long it shall be displayed.
type Frame struct {
	Pixels      []byte
	DisplayTime time.Duration
}
