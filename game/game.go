// Package game plugs the Tak engine into the Concord table kit: the rules
// adapter, the computer player, and the board people play on.
package game

import (
	"fmt"
	"strconv"

	"github.com/JMThomas00/Concord/sdk/table"
	"github.com/JMThomas00/concord-tak/engine"
	tea "github.com/charmbracelet/bubbletea"
)

// Option keys: the channel's settings in Concord, and flags standalone.
const (
	OptionSize = "board_size" // "3".."8", default "5"
	OptionKomi = "komi"       // points added to Black's flat count: "0", "0.5", ... "2.5"
)

// Rules is Tak for the table kit.
var Rules = table.Rules{
	Name:      "Tak",
	SeatNames: []string{"White", "Black"},
	New:       func(options map[string]string) table.Game { return New(options) },
	NewBoard:  func(s *table.Seat) tea.Model { return newBoard(s) },
	AI: func(g table.Game, level int) string {
		b := g.(*Game).B
		return engine.Best(b, level).PTN(b)
	},
}

// Game adapts an engine.Board to table.Game.
type Game struct {
	B    *engine.Board
	Last string // the last move, in PTN
}

// New starts a game with the given options (bad values fall back to 5×5,
// no komi).
func New(options map[string]string) *Game {
	size, err := strconv.Atoi(options[OptionSize])
	if err != nil || size < 3 || size > 8 {
		size = 5
	}
	komi2 := 0
	if k, err := strconv.ParseFloat(options[OptionKomi], 64); err == nil && k >= 0 && k <= 10 {
		komi2 = int(k * 2)
	}
	b, _ := engine.New(size, komi2)
	return &Game{B: b}
}

func (g *Game) Turn() int {
	if g.B.Result().Over {
		return -1
	}
	return g.B.ToMove
}

func (g *Game) Play(move string) error {
	if g.B.Result().Over {
		return fmt.Errorf("the game is over")
	}
	m, err := g.B.Parse(move)
	if err != nil {
		return err
	}
	g.Last = m.PTN(g.B)
	g.B = g.B.Apply(m)
	return nil
}

func (g *Game) Outcome() table.Outcome {
	r := g.B.Result()
	return table.Outcome{Over: r.Over, Winner: r.Winner, Reason: r.Reason}
}
