package game

import "github.com/JMThomas00/Concord/sdk/table"

// sound is what everyone watching hears after a move.
func sound(tg table.Game, move string) string {
	g := tg.(*Game)
	switch o := g.Outcome(); {
	case o.Over && o.Winner >= 0:
		return "sounds/finish.wav"
	case o.Over:
		return "sounds/draw.wav"
	case g.Crushed:
		return "sounds/crush.wav"
	case g.Slid:
		return "sounds/drop.wav"
	}
	return "sounds/clack.wav"
}
