package audio

import (
	"bytes"
	"io"
	"math"
	"os"
	"time"

	"github.com/ebitengine/oto/v3"
	"github.com/pudelkoM/go-render/pkg/blockworld"
)

type Audio struct {
	otoCtx  *oto.Context
	gig_wav []byte
}

func InitAudio() *Audio {
	ctx := &Audio{}

	gig_wav, err := os.Open("./assets/ehehe_trim.wav")
	if err != nil {
		panic("os.ReadFile failed: " + err.Error())
	}
	defer gig_wav.Close()

	ctx.gig_wav, err = io.ReadAll(gig_wav)
	if err != nil {
		panic("io.ReadAll failed: " + err.Error())
	}

	op := &oto.NewContextOptions{}

	// Usually 44100 or 48000. Other values might cause distortions in Oto
	op.SampleRate = 48000

	// Number of channels (aka locations) to play sounds from. Either 1 or 2.
	// 1 is mono sound, and 2 is stereo (most speakers are stereo).
	op.ChannelCount = 2

	// Format of the source. go-mp3's format is signed 16bit integers.
	op.Format = oto.FormatSignedInt16LE

	// Remember that you should **not** create more than one context
	otoCtx, readyChan, err := oto.NewContext(op)
	if err != nil {
		panic("oto.NewContext failed: " + err.Error())
	}
	// It might take a bit for the hardware audio devices to be ready, so we wait on the channel.
	<-readyChan

	ctx.otoCtx = otoCtx

	return ctx
}

func HandleAudio(ctx *Audio, world *blockworld.Blockworld, frameCount int64) {
	if frameCount%67 != 0 {
		return
	}

	r := bytes.NewReader(ctx.gig_wav)

	// Create a new 'player' that will handle our sound. Paused by default.
	player := ctx.otoCtx.NewPlayer(r)

	// Play starts playing the sound and returns without waiting for it (Play() is async).
	player.SetVolume(0.01)
	player.Play()

	go func() {
		// We can wait for the sound to finish playing using something like this
		for player.IsPlaying() {
			time.Sleep(time.Millisecond)
			player.SetVolume(math.Min(player.Volume()+0.01, 0.3))
		}
		player.Close()
	}()
}
