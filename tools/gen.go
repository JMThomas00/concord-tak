//go:build ignore

// gen writes Tak's sounds into client/sounds: its own (a stone placed, a
// stack sliding, a wall crushed, a draw) and the Concord Arcade kit. Run from the repo
// root: go run tools/gen.go
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"

	"github.com/JMThomas00/Concord/sdk/arcade"
)

const rate = 22050

func main() {
	check(arcade.WriteSoundKit("client"))
	sound("clack.wav", .09, func(t float64) float64 { // stone on stone
		return .5*math.Sin(2*math.Pi*1700*t)*math.Exp(-t*120) + .4*math.Sin(2*math.Pi*420*t)*math.Exp(-t*70)
	})
	sound("drop.wav", .3, func(t float64) float64 { // a stack set down along a line: three soft thunks
		v := 0.0
		for i, f := range []float64{520, 470, 420} {
			if s := t - float64(i)*.08; s >= 0 {
				v += .35 * math.Sin(2*math.Pi*f*s) * math.Exp(-s*45)
			}
		}
		return v
	})
	sound("crush.wav", .45, func(t float64) float64 { // a wall flattened: a crack and a low rumble
		f := 160 - 100*t
		return .45*sq(f, t)*math.Exp(-t*6) + .3*math.Sin(2*math.Pi*1300*t)*math.Exp(-t*90)
	})
	sound("draw.wav", .6, func(t float64) float64 { // two gentle notes, the second lower
		v := 0.0
		for i, f := range []float64{587.33, 493.88} {
			if s := t - float64(i)*.16; s >= 0 {
				v += .25 * math.Sin(2*math.Pi*f*s) * math.Exp(-s*6)
			}
		}
		return v
	})
	fmt.Println("tak sounds and the arcade kit written to client/sounds")
}

// sq is a square wave, softened.
func sq(f, t float64) float64 { return math.Tanh(4 * math.Sin(2*math.Pi*f*t)) }

func sound(name string, dur float64, f func(t float64) float64) {
	n := int(dur * rate)
	data := make([]byte, 2*n)
	for i := 0; i < n; i++ {
		t := float64(i) / rate
		fade := math.Min(1, (dur-t)/.01)
		v := max(-1, min(1, f(t)*fade))
		binary.LittleEndian.PutUint16(data[2*i:], uint16(int16(v*32000)))
	}
	var h bytes.Buffer
	h.WriteString("RIFF")
	binary.Write(&h, binary.LittleEndian, uint32(36+len(data)))
	h.WriteString("WAVEfmt ")
	for _, v := range []any{uint32(16), uint16(1), uint16(1), uint32(rate), uint32(rate * 2), uint16(2), uint16(16)} {
		binary.Write(&h, binary.LittleEndian, v)
	}
	h.WriteString("data")
	binary.Write(&h, binary.LittleEndian, uint32(len(data)))
	h.Write(data)
	path := filepath.Join("client", "sounds", name)
	check(os.MkdirAll(filepath.Dir(path), 0o755))
	check(os.WriteFile(path, h.Bytes(), 0o644))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
