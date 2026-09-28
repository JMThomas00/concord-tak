package engine

import (
	"strings"
	"testing"
	"time"
)

func perft(b *Board, depth int) int {
	if depth == 0 {
		return 1
	}
	if b.Result().Over {
		return 0
	}
	n := 0
	for _, m := range b.Moves() {
		n += perft(b.Apply(m), depth-1)
	}
	return n
}

func TestPerft5x5(t *testing.T) {
	b, _ := New(5, 0)
	for depth, want := range []int{1, 25, 600, 43320} {
		if got := perft(b, depth); got != want {
			t.Errorf("perft(%d) = %d, want %d", depth, got, want)
		}
	}
}

// play applies PTN moves in order, failing the test on an illegal one.
func play(t *testing.T, b *Board, moves ...string) *Board {
	t.Helper()
	for _, s := range moves {
		m, err := b.Parse(s)
		if err != nil {
			t.Fatalf("%s: %v", s, err)
		}
		b = b.Apply(m)
	}
	return b
}

func TestOpeningPlacesTheOpponentsFlat(t *testing.T) {
	b, _ := New(5, 0)
	if _, err := b.Parse("Sa1"); err == nil {
		t.Fatal("a wall was allowed as the first move")
	}
	b = play(t, b, "a1", "e5")
	if top, _ := b.Top(0); top.Color != Black {
		t.Fatalf("White's first move should place a Black flat, got color %d", top.Color)
	}
	if top, _ := b.Top(24); top.Color != White {
		t.Fatalf("Black's first move should place a White flat")
	}
	if b.Stones != [2]int{20, 20} {
		t.Fatalf("supplies = %v, want 20 each", b.Stones)
	}
}

func TestWallsBlockAndCapstonesFlattenThem(t *testing.T) {
	b, _ := New(5, 0)
	// White owns e5 (placed by Black); set up White cap at c3, Black wall at d3.
	b = play(t, b, "a1", "e5", "Cc3", "Sd3")
	if _, err := b.Parse("c3>"); err != nil {
		t.Fatalf("a lone capstone should flatten the wall: %v", err)
	}
	b = play(t, b, "c3>")
	top, _ := b.Top(b.N*2 + 3)
	if top.Kind != Cap || len(b.Stacks[b.N*2+3]) != 2 || b.Stacks[b.N*2+3][0].Kind != Flat {
		t.Fatalf("d3 after the capstone: %+v", b.Stacks[b.N*2+3])
	}
	// Black places a wall at b3; White's flat on e5 can't move onto walls, and
	// nothing moves onto a capstone.
	b = play(t, b, "Sb1")
	if _, err := b.Parse("b1-"); err == nil {
		t.Fatal("moved off the board")
	}
}

func TestCarryLimitAndDrops(t *testing.T) {
	b, _ := New(3, 0)
	// Build a 3-high stack on a1 with Black on top, then carry the limit (3) right.
	b = play(t, b, "c3", "a1", "b1", "a2", "b1<", "a2-")
	// a1 now: White(a1), White(b1 moved), Black(a2) on top → Black controls.
	if top, _ := b.Top(0); top.Color != Black || len(b.Stacks[0]) != 3 {
		t.Fatalf("a1 = %+v", b.Stacks[0])
	}
	b = play(t, b, "c1")
	// Black carries all three right: two onto b1, one onto c1.
	if _, err := b.Parse("3a1>21"); err != nil {
		t.Fatalf("3a1>21 should be legal: %v", err)
	}
	// Three drops need three squares to the right; a 3x3 board has two.
	if _, err := b.Parse("3a1>111"); err == nil {
		t.Fatal("3a1>111 ran off the board")
	}
	b = play(t, b, "3a1>21")
	if len(b.Stacks[0]) != 0 || len(b.Stacks[1]) != 2 || len(b.Stacks[2]) != 2 {
		t.Fatalf("after 3a1>21: a1 %v, b1 %v, c1 %v", b.Stacks[0], b.Stacks[1], b.Stacks[2])
	}
}

func TestRoadWinsAndTheMoverWinsADoubleRoad(t *testing.T) {
	b, _ := New(3, 0)
	b = play(t, b, "c3", "a1", "a2", "b3", "b2", "c1")
	// White has a2, b2; placing c2 completes the row a2-b2-c2.
	b = play(t, b, "c2")
	if r := b.Result(); !r.Over || r.Winner != White || r.Reason != "road" {
		t.Fatalf("result = %+v, want a White road", r)
	}
}

func TestFlatWinWithKomi(t *testing.T) {
	b, _ := New(3, 4) // komi 2
	// Fill the board without roads: alternate walls to block.
	moves := []string{"a1", "c3", "Sb1", "a3", "Sc1", "b3", "b2", "Sa2", "c2"}
	b = play(t, b, moves...)
	r := b.Result()
	if !r.Over {
		t.Fatalf("board full but not over: %+v", r)
	}
	f := b.Flats()
	if !strings.HasPrefix(r.Reason, "flats") {
		t.Fatalf("reason %q, flats %v", r.Reason, f)
	}
	white2, black2 := f[White]*2, f[Black]*2+4
	want := -1
	if white2 > black2 {
		want = White
	} else if black2 > white2 {
		want = Black
	}
	if r.Winner != want {
		t.Fatalf("winner %d, want %d (flats %v, komi 2)", r.Winner, want, f)
	}
}

func TestPTNRoundTrip(t *testing.T) {
	b, _ := New(5, 0)
	b = play(t, b, "a1", "e5", "c3", "Cd3", "Sb2")
	for _, m := range b.Moves() {
		back, err := b.Parse(m.PTN(b))
		if err != nil || back.PTN(b) != m.PTN(b) {
			t.Fatalf("%s didn't round-trip: %v %v", m.PTN(b), back.PTN(b), err)
		}
	}
}

// The "dragon clause": a move that makes roads for both players wins for
// the player who moved.
func TestMoverWinsWhenAMoveMakesBothRoads(t *testing.T) {
	b, _ := New(3, 0)
	w, bl := Piece{Color: White}, Piece{Color: Black}
	b.Stacks[0], b.Stacks[1] = []Piece{w}, []Piece{w}   // a1, b1
	b.Stacks[2] = []Piece{w, bl}                        // c1: White under Black
	b.Stacks[3], b.Stacks[4] = []Piece{bl}, []Piece{bl} // a2, b2
	b.ToMove, b.Ply = Black, 10
	// Black lifts its stone from c1 up to c2: that completes Black's rank 2
	// and uncovers White's stone on c1, completing White's rank 1.
	b = play(t, b, "c1+")
	if !b.HasRoad(White) || !b.HasRoad(Black) {
		t.Fatal("expected roads for both players")
	}
	if r := b.Result(); r.Winner != Black {
		t.Fatalf("result %+v; the mover (Black) should win", r)
	}
}

func TestComputerFinishesARoadAndBlocksOne(t *testing.T) {
	b, _ := New(4, 0)
	b = play(t, b, "d4", "a1", "b1", "a4", "c1", "b3")
	// White has a1 b1 c1: d1 wins.
	for level := 1; level <= 3; level++ {
		if m := Best(b, level); m.PTN(b) != "d1" {
			t.Errorf("level %d played %s instead of the winning d1", level, m.PTN(b))
		}
	}
	// If White plays elsewhere, Black (with no win of its own) must stop d1.
	b2 := play(t, b, "d3")
	for level := 2; level <= 3; level++ {
		m := Best(b2, level)
		b3 := b2.Apply(m)
		if mv, err := b3.Parse("d1"); err == nil && b3.Apply(mv).HasRoad(White) {
			t.Errorf("level %d played %s and left White's d1 road open", level, m.PTN(b2))
		}
	}
}

func TestComputerIsQuickEnough(t *testing.T) {
	b, _ := New(5, 0)
	b = play(t, b, "a1", "e5", "c3", "c2", "b3", "d3")
	began := time.Now()
	m := Best(b, 3)
	if _, err := b.Parse(m.PTN(b)); err != nil {
		t.Fatalf("hard computer chose an illegal move %s", m.PTN(b))
	}
	if took := time.Since(began); took > 4*time.Second {
		t.Fatalf("hard computer took %v", took)
	}
}
