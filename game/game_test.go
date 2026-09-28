package game

import (
	"context"
	"testing"
	"time"

	"github.com/JMThomas00/Concord/sdk/plugin"
	"github.com/JMThomas00/Concord/sdk/plugintest"
	"github.com/JMThomas00/Concord/sdk/table"
	"github.com/JMThomas00/Concord/sdk/wire"
	"github.com/google/uuid"
)

func TestGameAdapter(t *testing.T) {
	g := Rules.New(map[string]string{OptionSize: "4", OptionKomi: "1.5"}).(*Game)
	if g.B.N != 4 || g.B.Komi2 != 3 || g.Turn() != 0 {
		t.Fatalf("size %d, komi2 %d, turn %d", g.B.N, g.B.Komi2, g.Turn())
	}
	if d := Rules.New(map[string]string{OptionSize: "12"}).(*Game); d.B.N != 5 || d.B.Komi2 != 0 {
		t.Fatalf("bad options should give 5×5 without komi; got %d, %d", d.B.N, d.B.Komi2)
	}
	if err := g.Play("Sa1"); err == nil {
		t.Fatal("a wall was allowed as the first move")
	}
	for _, m := range []string{"d4", "a1", "b1", "c4", "c1"} {
		if err := g.Play(m); err != nil {
			t.Fatalf("%s: %v", m, err)
		}
	}
	if g.Last != "c1" || g.Turn() != 1 {
		t.Fatalf("last %q, turn %d", g.Last, g.Turn())
	}
	for _, m := range []string{"b4", "a2", "a4"} {
		if err := g.Play(m); err != nil {
			t.Fatalf("%s: %v", m, err)
		}
	}
	// Black now has a4 b4 c4 d4: a road.
	if o := g.Outcome(); !o.Over || o.Winner != 1 || g.Turn() != -1 {
		t.Fatalf("outcome %+v, turn %d", o, g.Turn())
	}
	if err := g.Play("d1"); err == nil {
		t.Fatal("played on after the game ended")
	}
}

func TestComputerPlaysLegalMoves(t *testing.T) {
	for level := 1; level <= 3; level++ {
		g := Rules.New(map[string]string{OptionSize: "4"})
		began := time.Now()
		for i := 0; i < 4 && !g.Outcome().Over; i++ {
			mv := Rules.AI(g, level)
			if err := g.Play(mv); err != nil {
				t.Fatalf("level %d suggested an illegal move %q: %v", level, mv, err)
			}
		}
		if took := time.Since(began); took > 15*time.Second {
			t.Errorf("level %d took %v for 4 moves", level, took)
		}
	}
}

// Two players sit down, place stones with the keyboard and by typing PTN,
// and move a stack.
func TestTwoPlayersInAChannel(t *testing.T) {
	srv := plugintest.NewServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go plugin.Run(ctx, srv.Config(), table.New(Rules).Handler())
	srv.WaitReady()
	ch := uuid.New()
	srv.Channel(wire.Channel{ID: ch, Name: "tak", PluginConfig: map[string]string{OptionSize: "4"}})

	white := srv.Enter(ch, "alice", 60, 24)
	black := srv.Enter(ch, "bob", 60, 24)
	srv.FrameContaining(white, "Tab: sit down")
	srv.FrameContaining(black, "Tab: sit down")
	for _, v := range []*plugintest.Viewer{white, black} {
		srv.Key(v, "tab")
		srv.Key(v, "enter")
	}
	srv.FrameContaining(white, "place your opponent's first stone")

	// White's cursor starts on c3: that places Black's first stone.
	srv.Key(white, "enter")
	srv.FrameContaining(black, "place your opponent's first stone")
	srv.Key(black, "enter") // c3 is taken
	srv.FrameContaining(black, "first, place your opponent's stone on an empty square")
	srv.Key(black, "left")
	srv.Key(black, "enter") // b3: White's first stone
	srv.FrameContaining(white, "your move")

	// White types a move.
	srv.Key(white, ":")
	srv.FrameContaining(white, "move (PTN")
	srv.Type(white, "Sa1")
	srv.Key(white, "enter")
	srv.FrameContaining(white, "last move Sa1")

	// Black picks up its stone on c3 and moves it down to c2.
	srv.Key(black, "right")
	srv.Key(black, "enter")
	srv.FrameContaining(black, "carrying 1")
	srv.Key(black, "down")
	srv.FrameContaining(black, "last move c3-")
	t.Log("\n" + srv.FrameContaining(white, "your move"))
}
