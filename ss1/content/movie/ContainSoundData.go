package movie

import (
	"bytes"

	"github.com/inkyblackness/hacked/ss1/content/audio"
)

// ContainSoundData packs a sound data into a container and encodes it.
func ContainSoundData(soundData audio.L8) []byte {
	var container Container
	startOffset := 0

	for (startOffset + audioEntrySize) <= len(soundData.Samples) {
		ts := TimestampFromSeconds(float32(startOffset) / soundData.SampleRate)
		endOffset := startOffset + audioEntrySize
		container.AddEntry(Entry{
			Timestamp: ts,
			Data: AudioEntryData{
				Samples: soundData.Samples[startOffset:endOffset],
			},
		})
		startOffset = endOffset
	}
	if startOffset < len(soundData.Samples) {
		ts := TimestampFromSeconds(float32(startOffset) / soundData.SampleRate)
		container.AddEntry(Entry{
			Timestamp: ts,
			Data: AudioEntryData{
				Samples: soundData.Samples[startOffset:],
			},
		})
	}

	container.EndTimestamp = TimestampFromSeconds(float32(len(soundData.Samples)) / soundData.SampleRate)
	container.Audio.Sound.SampleRate = soundData.SampleRate

	buffer := bytes.NewBuffer(nil)
	_ = Write(buffer, container)
	return buffer.Bytes()
}
