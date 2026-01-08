package audio

type WavData struct {
	Metadata      WavMetadata
	Samples       []byte
	ChunkID       int
	cursorSamples int
}

func (wd *WavData) Advance(n int) (chunk []byte, eof bool) {
	if n <= 0 {
		return nil, wd.isEOF()
	}

	frameSize := wd.bytesPerFrame()
	totalFrames := len(wd.Samples) / frameSize

	if wd.cursorSamples >= totalFrames {
		return nil, true
	}

	end := min(wd.cursorSamples+n, totalFrames)

	startByte := wd.cursorSamples * frameSize
	endByte := end * frameSize

	chunk = wd.Samples[startByte:endByte]
	wd.cursorSamples = end
	wd.ChunkID++
	return chunk, wd.cursorSamples >= totalFrames
}

func (wd *WavData) Reset() {
	wd.cursorSamples = 0
}

func (wd *WavData) RemainingSamples() int {
	bps := wd.bytesPerFrame()
	totalSamples := len(wd.Samples) / bps
	if wd.cursorSamples >= totalSamples {
		return 0
	}
	return totalSamples - wd.cursorSamples
}

func (wd *WavData) isEOF() bool {
	bps := wd.bytesPerFrame()
	return wd.cursorSamples >= (len(wd.Samples) / bps)
}

func (wd *WavData) bytesPerFrame() int {
	return int(wd.Metadata.Channels) * int(wd.Metadata.Bitdepth/8)
}
