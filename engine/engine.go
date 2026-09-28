// Package engine is Tak, following the standard (US Tak Association) rules.
//
//   - Boards are 3×3 to 8×8. Each player has a fixed supply of stones (and,
//     on 5×5 and up, capstones): 3→10/0, 4→15/0, 5→21/1, 6→30/1, 7→40/2,
//     8→50/2.
//   - On their first turn, each player places one of the *opponent's* flat
//     stones on an empty square.
//   - A turn is either placing a stone from your supply on an empty square
//     (as a flat, a standing stone/"wall", or a capstone), or moving a stack
//     you control (your stone on top): pick up to the board size of its top
//     stones, move them in a straight line, dropping at least one on each
//     square you pass. Nothing may move onto a wall or capstone, except a
//     capstone moving alone onto a wall, which flattens it.
//   - A road -- a connected line of your flats and capstones (not walls)
//     joining opposite edges -- wins. If a move makes roads for both
//     players, the player who moved wins.
//   - If the board fills, or either player plays their last stone, the
//     player with more flats on top wins (walls and capstones don't count);
//     komi, if set, is added to the second player's count. Equal is a draw.
//
// Moves are written in Portable Tak Notation (PTN): "a1" places a flat,
// "Sa1" a wall, "Ca1" a capstone; "3c3>12" picks up 3 stones on c3 and
// moves right, dropping 1 then 2. Files run a.. left to right, ranks 1..
// bottom to top; directions are + (up), - (down), > (right), < (left).
package engine

import (
	"fmt"
	"strconv"
	"strings"
)

// Colors: White moves first.
const (
	White = 0
	Black = 1
)

// Kinds of stone.
const (
	Flat = iota
	Wall
	Cap
)

// Piece is one stone.
type Piece struct {
	Color int8
	Kind  int8
}

// Supply is how many stones and capstones each player starts with.
var Supply = map[int][2]int{3: {10, 0}, 4: {15, 0}, 5: {21, 1}, 6: {30, 1}, 7: {40, 2}, 8: {50, 2}}

// Board is a position.
type Board struct {
	N      int
	Stacks [][]Piece // index rank*N + file; each stack bottom → top
	ToMove int
	Ply    int
	Stones [2]int
	Caps   [2]int
	Komi2  int // komi in half points (komi 2.5 → 5), added to Black's flat count
}

// New starts an empty n×n game; komi is in half points.
func New(n, komi2 int) (*Board, error) {
	s, ok := Supply[n]
	if !ok {
		return nil, fmt.Errorf("board size must be 3 to 8")
	}
	return &Board{N: n, Stacks: make([][]Piece, n*n), Stones: [2]int{s[0], s[0]}, Caps: [2]int{s[1], s[1]}, Komi2: komi2}, nil
}

// Clone copies b deeply.
func (b *Board) Clone() *Board {
	nb := *b
	nb.Stacks = make([][]Piece, len(b.Stacks))
	for i, st := range b.Stacks {
		if len(st) > 0 {
			nb.Stacks[i] = append([]Piece(nil), st...)
		}
	}
	return &nb
}

// Top returns the top stone of a square, if any.
func (b *Board) Top(sq int) (Piece, bool) {
	st := b.Stacks[sq]
	if len(st) == 0 {
		return Piece{}, false
	}
	return st[len(st)-1], true
}

// Name is a square's name, e.g. "c3".
func (b *Board) Name(sq int) string {
	return string(rune('a'+sq%b.N)) + strconv.Itoa(sq/b.N+1)
}

// Direction deltas: up, down, right, left.
var dirChars = []byte{'+', '-', '>', '<'}

func (b *Board) step(sq, dir int) int {
	file, rank := sq%b.N, sq/b.N
	switch dir {
	case 0:
		rank++
	case 1:
		rank--
	case 2:
		file++
	case 3:
		file--
	}
	if file < 0 || file >= b.N || rank < 0 || rank >= b.N {
		return -1
	}
	return rank*b.N + file
}

// Move is a placement (Drops empty) or a slide.
type Move struct {
	Sq    int
	Kind  int8  // placement: Flat, Wall or Cap
	Dir   int   // slide: 0 up, 1 down, 2 right, 3 left
	Drops []int // slide: stones dropped on each square in turn
}

// IsSlide reports whether m moves a stack.
func (m Move) IsSlide() bool { return len(m.Drops) > 0 }

// Carry is how many stones a slide picks up.
func (m Move) Carry() int {
	n := 0
	for _, d := range m.Drops {
		n += d
	}
	return n
}

// PTN formats m in Portable Tak Notation on board b.
func (m Move) PTN(b *Board) string {
	if !m.IsSlide() {
		return [...]string{"", "S", "C"}[m.Kind] + b.Name(m.Sq)
	}
	var sb strings.Builder
	if c := m.Carry(); c > 1 {
		sb.WriteString(strconv.Itoa(c))
	}
	sb.WriteString(b.Name(m.Sq))
	sb.WriteByte(dirChars[m.Dir])
	if len(m.Drops) > 1 {
		for _, d := range m.Drops {
			sb.WriteString(strconv.Itoa(d))
		}
	}
	return sb.String()
}

// Moves lists every legal move.
func (b *Board) Moves() []Move {
	var out []Move
	me := b.ToMove
	if b.Ply < 2 { // opening: place an opponent's flat
		for sq, st := range b.Stacks {
			if len(st) == 0 {
				out = append(out, Move{Sq: sq, Kind: Flat})
			}
		}
		return out
	}
	for sq, st := range b.Stacks {
		if len(st) == 0 {
			if b.Stones[me] > 0 {
				out = append(out, Move{Sq: sq, Kind: Flat}, Move{Sq: sq, Kind: Wall})
			}
			if b.Caps[me] > 0 {
				out = append(out, Move{Sq: sq, Kind: Cap})
			}
			continue
		}
		if int(st[len(st)-1].Color) != me {
			continue
		}
		maxCarry := len(st)
		if maxCarry > b.N {
			maxCarry = b.N
		}
		capOnTop := st[len(st)-1].Kind == Cap
		for dir := 0; dir < 4; dir++ {
			for carry := 1; carry <= maxCarry; carry++ {
				b.slides(sq, dir, carry, capOnTop, nil, sq, &out)
			}
		}
	}
	return out
}

// slides extends a slide from `at` carrying `left` stones, collecting every
// legal way to drop them.
func (b *Board) slides(from, dir, left int, capOnTop bool, drops []int, at int, out *[]Move) {
	next := b.step(at, dir)
	if next < 0 {
		return
	}
	if top, ok := b.Top(next); ok {
		switch top.Kind {
		case Cap:
			return
		case Wall:
			// Only a capstone arriving alone may flatten a wall, as the last drop.
			if !(capOnTop && left == 1) {
				return
			}
			*out = append(*out, Move{Sq: from, Dir: dir, Drops: append(append([]int(nil), drops...), 1)})
			return
		}
	}
	for drop := 1; drop <= left; drop++ {
		d := append(append([]int(nil), drops...), drop)
		if drop == left {
			*out = append(*out, Move{Sq: from, Dir: dir, Drops: d})
		} else {
			b.slides(from, dir, left-drop, capOnTop, d, next, out)
		}
	}
}

// Apply plays m (from Moves) and returns the new position.
func (b *Board) Apply(m Move) *Board {
	nb := b.Clone()
	me := b.ToMove
	if !m.IsSlide() {
		color := me
		if b.Ply < 2 {
			color = 1 - me // opening: the opponent's flat, from their supply
		}
		if m.Kind == Cap {
			nb.Caps[color]--
		} else {
			nb.Stones[color]--
		}
		nb.Stacks[m.Sq] = append(nb.Stacks[m.Sq], Piece{Color: int8(color), Kind: m.Kind})
	} else {
		st := nb.Stacks[m.Sq]
		carry := m.Carry()
		hand := append([]Piece(nil), st[len(st)-carry:]...)
		nb.Stacks[m.Sq] = st[:len(st)-carry]
		at := m.Sq
		for _, d := range m.Drops {
			at = nb.step(at, m.Dir)
			if top, ok := nb.Top(at); ok && top.Kind == Wall {
				nb.Stacks[at][len(nb.Stacks[at])-1].Kind = Flat // flattened by a capstone
			}
			nb.Stacks[at] = append(nb.Stacks[at], hand[:d]...)
			hand = hand[d:]
		}
	}
	nb.ToMove = 1 - me
	nb.Ply++
	return nb
}

// Parse reads a PTN move and returns the matching legal move.
func (b *Board) Parse(s string) (Move, error) {
	s = strings.TrimRight(strings.TrimSpace(s), "*'!?\"")
	if s == "" {
		return Move{}, fmt.Errorf("empty move")
	}
	var want string
	lower := strings.ToLower(s)
	if i := strings.IndexAny(lower, "+-<>"); i >= 0 { // slide
		count, rest := 1, lower
		if lower[0] >= '1' && lower[0] <= '9' {
			count, rest = int(lower[0]-'0'), lower[1:]
			i--
		}
		if len(rest) < 3 {
			return Move{}, fmt.Errorf("%q isn't a move", s)
		}
		sq, err := b.square(rest[:i])
		if err != nil {
			return Move{}, err
		}
		dir := strings.IndexByte(string(dirChars), rest[i])
		dropStr := rest[i+1:]
		var drops []int
		if dropStr == "" {
			drops = []int{count}
		} else {
			for _, c := range dropStr {
				if c < '1' || c > '9' {
					return Move{}, fmt.Errorf("%q isn't a move", s)
				}
				drops = append(drops, int(c-'0'))
			}
		}
		want = Move{Sq: sq, Dir: dir, Drops: drops}.PTN(b)
	} else { // placement
		kind := int8(Flat)
		name := lower
		switch s[0] {
		case 'S':
			kind, name = Wall, lower[1:]
		case 'C':
			kind, name = Cap, lower[1:]
		case 'F':
			name = lower[1:]
		}
		sq, err := b.square(name)
		if err != nil {
			return Move{}, err
		}
		want = Move{Sq: sq, Kind: kind}.PTN(b)
	}
	for _, m := range b.Moves() {
		if m.PTN(b) == want {
			return m, nil
		}
	}
	if b.Ply < 2 && !strings.ContainsAny(lower, "+-<>") && want != strings.TrimPrefix(want, "S") {
		return Move{}, fmt.Errorf("the first move of each player must be a flat")
	}
	return Move{}, fmt.Errorf("%s isn't a legal move here", s)
}

func (b *Board) square(name string) (int, error) {
	if len(name) < 2 {
		return 0, fmt.Errorf("%q isn't a square", name)
	}
	file := int(name[0] - 'a')
	rank, err := strconv.Atoi(name[1:])
	if err != nil || file < 0 || file >= b.N || rank < 1 || rank > b.N {
		return 0, fmt.Errorf("%q isn't a square on this board", name)
	}
	return (rank-1)*b.N + file, nil
}

// HasRoad reports whether color has a road.
func (b *Board) HasRoad(color int) bool {
	n := b.N
	road := func(sq int) bool {
		top, ok := b.Top(sq)
		return ok && int(top.Color) == color && top.Kind != Wall
	}
	// Two searches: left edge → right edge, and bottom edge → top edge.
	for _, vertical := range []bool{false, true} {
		seen := make([]bool, n*n)
		var queue []int
		for i := 0; i < n; i++ {
			sq := i * n // left edge
			if vertical {
				sq = i // bottom edge
			}
			if road(sq) {
				seen[sq], queue = true, append(queue, sq)
			}
		}
		for len(queue) > 0 {
			sq := queue[0]
			queue = queue[1:]
			if (!vertical && sq%n == n-1) || (vertical && sq/n == n-1) {
				return true
			}
			for dir := 0; dir < 4; dir++ {
				if nx := b.step(sq, dir); nx >= 0 && !seen[nx] && road(nx) {
					seen[nx], queue = true, append(queue, nx)
				}
			}
		}
	}
	return false
}

// Flats counts flats on top for each color.
func (b *Board) Flats() [2]int {
	var f [2]int
	for sq := range b.Stacks {
		if top, ok := b.Top(sq); ok && top.Kind == Flat {
			f[top.Color]++
		}
	}
	return f
}

// Result says whether the game is over after the last move.
type Result struct {
	Over   bool
	Winner int // White, Black, or -1 for a draw
	Reason string
}

// Result checks roads, then the flat count.
func (b *Board) Result() Result {
	if b.Ply == 0 {
		return Result{}
	}
	mover := 1 - b.ToMove
	w, bl := b.HasRoad(White), b.HasRoad(Black)
	switch {
	case w && bl:
		return Result{Over: true, Winner: mover, Reason: "road"}
	case w:
		return Result{Over: true, Winner: White, Reason: "road"}
	case bl:
		return Result{Over: true, Winner: Black, Reason: "road"}
	}
	full := true
	for _, st := range b.Stacks {
		if len(st) == 0 {
			full = false
			break
		}
	}
	out := func(c int) bool { return b.Stones[c]+b.Caps[c] == 0 }
	if !full && !out(White) && !out(Black) {
		return Result{}
	}
	f := b.Flats()
	white2, black2 := f[White]*2, f[Black]*2+b.Komi2
	reason := fmt.Sprintf("flats %d–%s", f[White], halves(black2))
	switch {
	case white2 > black2:
		return Result{Over: true, Winner: White, Reason: reason}
	case black2 > white2:
		return Result{Over: true, Winner: Black, Reason: reason}
	}
	return Result{Over: true, Winner: -1, Reason: reason}
}

// halves formats a half-point count ("10", "10.5").
func halves(h int) string {
	if h%2 == 0 {
		return strconv.Itoa(h / 2)
	}
	return fmt.Sprintf("%d.5", h/2)
}
