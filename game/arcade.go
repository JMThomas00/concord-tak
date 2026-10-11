package game

import (
	"fmt"
	"strconv"

	"github.com/JMThomas00/Concord/sdk/arcade"
	"github.com/JMThomas00/Concord/sdk/table"
	"github.com/JMThomas00/concord-tak/engine"
)

// The arcade front door's personality: the stone garden. Stacks are drawn
// side on, a winning road lights up edge to edge, a capstone flattening a
// wall gets a CRUSH!, and Pebbles buy stone sets and boards at THE QUARRY.
// The board size and komi are chosen on the NEW GAME screen for each game.
var arcadeLook = &table.Arcade{
	Title:   "TAK",
	Tagline: "BUILD A ROAD",
	HowTo: []string{
		"Build a road: your flats and capstones joining opposite edges.",
		"Each turn place a stone from your supply on an empty square, or",
		"move a stack you control (your stone on top).",
		"A WALL blocks roads; nothing lands on it, except a CAPSTONE moving",
		"alone, which flattens it: CRUSH! Pick up to as many stones as the",
		"board is wide; drop at least one on each square, in a straight line.",
		"Your first move places one of your opponent's flats.",
		"Board full or stones gone: most flats on top wins (Black gets komi).",
	},
	Keys: []arcade.Key{
		{Key: "Enter", Does: "flat, or pick up"},
		{Key: "S C", Does: "wall, capstone"},
		{Key: "+ -", Does: "carry more, fewer"},
		{Key: "Space", Does: "drop another here"},
		{Key: ":", Does: "type a move (c3, Sd4, 3c3>12)"},
	},
	Sounds:      true,
	Reward:      "PEBBLE",
	Collection:  "THE QUARRY",
	Unlockables: unlockables(),
	Kinds:       []table.Kind{{ID: kindStones, Label: "STONES"}, {ID: kindBoard, Label: "BOARD"}},
	Preview:     preview,
	Attract:     attract,
	Result:      result,
	Options: []table.Option{
		{Key: OptionSize, Label: "BOARD", Values: []string{"3", "4", "5", "6", "7", "8"},
			Names: []string{"3x3", "4x4", "5x5", "6x6", "7x7", "8x8"}, Default: "5",
			Describe: func(v string) string {
				n, _ := strconv.Atoi(v)
				s := engine.Supply[n]
				switch s[1] {
				case 0:
					return fmt.Sprintf("%d STONES EACH", s[0])
				case 1:
					return fmt.Sprintf("%d STONES, 1 CAPSTONE EACH", s[0])
				}
				return fmt.Sprintf("%d STONES, %d CAPSTONES EACH", s[0], s[1])
			}},
		{Key: OptionKomi, Label: "KOMI", Values: []string{"0", "0.5", "1", "1.5", "2", "2.5"},
			Names: []string{"KOMI 0", "KOMI 0.5", "KOMI 1", "KOMI 1.5", "KOMI 2", "KOMI 2.5"}, Default: "0",
			Describe: func(v string) string {
				if v == "0" {
					return "NO HANDICAP"
				}
				return "BLACK GETS " + v + " IN A FLAT COUNT"
			}},
	},
	SetupPreview: setupPreview,
}

func result(g table.Game, o table.Outcome, seats []string) string {
	switch {
	case o.Winner < 0:
		return "DRAW!"
	case o.Reason == "road":
		return "ROAD WIN!"
	}
	return "FLAT WIN!"
}
