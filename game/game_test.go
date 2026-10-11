package game

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JMThomas00/Concord/sdk/arcade"
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
	for _, v := range []*plugintest.Viewer{white, black} { // too small for the arcade: the plain door
		srv.FrameContaining(v, "Enter to play")
		srv.Key(v, "enter")
		srv.FrameContaining(v, "M: sit down")
	}
	for _, v := range []*plugintest.Viewer{white, black} {
		srv.Key(v, "m")
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
	// The prompt claims Esc (Esc cancels it) and takes every key, M included.
	if srv.Key(white, "esc") {
		t.Fatal("Esc was claimed with nothing to cancel")
	}
	srv.Key(white, ":")
	srv.FrameContaining(white, "move (PTN")
	if !srv.Key(white, "esc") {
		t.Fatal("the move prompt didn't claim Esc")
	}
	srv.FrameContaining(white, "your move")
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

func startServer(t *testing.T) (*plugintest.Server, uuid.UUID) {
	t.Helper()
	srv := plugintest.NewServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { plugin.Run(ctx, srv.Config(), table.New(Rules).Handler()); close(done) }()
	t.Cleanup(func() { cancel(); <-done })
	srv.WaitReady()
	ch := uuid.New()
	srv.Channel(wire.Channel{ID: ch, Name: "tak"})
	return srv, ch
}

// The board size is chosen on NEW GAME, for this game only.
func TestNewGameChoosesTheSize(t *testing.T) {
	srv, ch := startServer(t)
	v := srv.Enter(ch, "alice", 80, 24)
	srv.FrameContaining(v, "PRESS ENTER")
	srv.Key(v, "enter")
	srv.FrameContaining(v, "THE QUARRY")
	srv.Key(v, "enter") // 1 PLAYER VS CPU
	frame := srv.FrameContaining(v, "◂ 5x5 ▸")
	if !strings.Contains(frame, "21 STONES, 1 CAPSTONE EACH") || !strings.Contains(frame, "◂ KOMI 0 ▸") {
		t.Fatalf("NEW GAME:\n%s", frame)
	}
	srv.Key(v, "up")
	srv.Key(v, "up") // BOARD
	srv.Key(v, "right")
	srv.FrameContaining(v, "30 STONES, 1 CAPSTONE EACH")
	srv.Key(v, "enter")
	frame = srv.FrameContaining(v, "VS CPU · NORMAL · 6x6 · KOMI 0")
	for _, want := range []string{"STACK", "EMPTY", "30 STONES", "FIRST MOVE: PLACE ONE OF YOUR OPPONENT'S FLATS"} {
		if !strings.Contains(frame, want) {
			t.Fatalf("the table is missing %q:\n%s", want, frame)
		}
	}
}

// Two players at the arcade table: the first moves, then carrying a stack,
// with the panel showing what's in hand.
func TestArcadeTable(t *testing.T) {
	srv, ch := startServer(t)
	white, black := srv.Enter(ch, "alice", 80, 24), srv.Enter(ch, "bob", 80, 24)
	for i, v := range []*plugintest.Viewer{white, black} {
		srv.FrameContaining(v, "PRESS ENTER")
		srv.Key(v, "enter")
		srv.FrameContaining(v, "TAKE A SEAT")
		srv.Key(v, "down")
		srv.Key(v, "enter")
		if i == 0 { // the first to sit chooses the game
			srv.FrameContaining(v, "TAKE THE WHITE SEAT")
			srv.Key(v, "enter")
		}
	}
	srv.FrameContaining(white, "FIRST MOVE")
	srv.Key(white, "enter") // c3: Black's stone
	srv.FrameContaining(black, "FIRST MOVE")
	srv.Key(black, "left")
	srv.Key(black, "enter") // b3: White's stone
	srv.FrameContaining(white, "ENTER A FLAT")
	srv.Key(white, "down")
	srv.Key(white, "enter") // c2
	srv.FrameContaining(black, "ENTER A FLAT")
	srv.Key(black, "right")
	if f := srv.FrameContaining(black, "1 HIGH · BLACK"); !strings.Contains(f, "STACK") {
		t.Fatalf("the panel:\n%s", f)
	}
	srv.Key(black, "enter") // pick up c3
	if f := srv.FrameContaining(black, "CARRYING 1"); !strings.Contains(f, "IN HAND") {
		t.Fatalf("carrying:\n%s", f)
	}
	srv.Key(black, "up") // drop it on c4
	srv.FrameContaining(white, "ENTER A FLAT")
}

// A capstone moving alone onto a wall flattens it: CRUSH!, with its sound.
func TestCrush(t *testing.T) {
	g := Rules.New(nil).(*Game)
	for _, m := range []string{"a1", "e5", "Cc3", "Sd3"} {
		if err := g.Play(m); err != nil {
			t.Fatal(m, err)
		}
	}
	if err := g.Play("c3>"); err != nil {
		t.Fatal(err)
	}
	if !g.Crushed || sound(g, "c3>") != "sounds/crush.wav" {
		t.Fatalf("crushed %v, sound %s", g.Crushed, sound(g, "c3>"))
	}
	if err := g.Play("b2"); err != nil {
		t.Fatal(err)
	}
	if g.Crushed || sound(g, "b2") != "sounds/clack.wav" {
		t.Fatal("a placement counted as a crush")
	}
}

// Every set, board and size draws inside its space, and attract mode is
// the same picture for the same frame.
func TestDrawing(t *testing.T) {
	sgr := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	const ax, ay = 10, 3
	inside := func(name string, size [2]int, c *arcade.Canvas) {
		for y, raw := range strings.Split(c.String(), "\n") {
			in := y >= ay && y < ay+size[1]
			for x, r := range []rune(sgr.ReplaceAllString(raw, "")) {
				if r != ' ' && (!in || x < ax || x >= ax+size[0]) {
					t.Fatalf("%s at %v: %q at %d,%d, outside its space", name, size, r, x, y)
				}
			}
			if !in && strings.Contains(raw, "\x1b[") {
				t.Fatalf("%s at %v: colour on row %d, outside its space", name, size, y)
			}
		}
	}
	for _, st := range boardStyles {
		for _, size := range [][2]int{{42, 18}, {40, 17}, {40, 12}, {30, 13}, {21, 6}, {20, 6}} {
			c := arcade.New(80, 24, arcade.NewPalette(nil))
			sel := map[string]string{kindStones: "pancake", kindBoard: boardPrefix + st.ID}
			preview(c, sel[kindBoard], sel, ax, ay, size[0], size[1], false)
			preview(c, "pancake", sel, ax, ay, size[0], size[1], true)
			inside(st.ID, size, c)
		}
	}
	for n := 3; n <= 8; n++ {
		c := arcade.New(80, 24, arcade.NewPalette(nil))
		setupPreview(c, map[string]string{OptionSize: strconv.Itoa(n)}, ax, ay, 30, 13)
		inside("setup "+strconv.Itoa(n), [2]int{30, 13}, c)
	}
	draw := func(frame int) string {
		c := arcade.New(80, 24, arcade.NewPalette(nil))
		attract(c, 2, 6, 76, 14, frame)
		return c.String()
	}
	for f := 0; f < 200; f += 7 {
		if draw(f) != draw(f) {
			t.Fatalf("attract mode changes within frame %d", f)
		}
	}
	if draw(0) == draw(30) {
		t.Fatal("attract mode doesn't move")
	}
}
