package game

import (
	"strconv"
	"time"

	"github.com/JMThomas00/Concord/sdk/arcade"
	"github.com/JMThomas00/concord-tak/engine"
)

// Drawing on the arcade canvas. Each square shows its stack side on,
// bottom to top: a flat is a slab one pixel tall, a wall stands up, a
// capstone is a dome, and the height is written in the corner. Rank 1 is at
// the bottom, as Tak boards are drawn.

// look is how to draw one board.
type look struct {
	set     *stoneSet
	style   *boardStyle
	cursor  int          // square under the cursor, or -1
	path    map[int]bool // a stack being carried: where it's been
	landing int          // ...and the square it's dropping on next, or -1
	last    map[int]bool // the last move's squares
	road    []int        // the winning road, edge to edge
	lit     int          // how many of its squares are lit
	ghost   bool         // the stones are a locked silhouette
	grey    bool         // the board is locked: grey squares
	coords  bool         // file letters under the board
}

// cellFor picks a square's size for an n x n board in w x h cells.
func cellFor(n, w, h int) (cw, ch int) {
	return max(1, min(12, w/n)), max(1, min(5, h/n))
}

// drawBoard draws position bd centred in w x h cells at (x, y).
func drawBoard(c *arcade.Canvas, x, y, w, h int, bd *engine.Board, l look) {
	n := bd.N
	if l.coords {
		h--
	}
	cw, ch := cellFor(n, w, h)
	x += max(0, (w-cw*n)/2)
	y += max(0, (h-ch*n)/2)
	lit := map[int]bool{}
	for i, sq := range l.road {
		if i < l.lit {
			lit[sq] = true
		}
	}
	for r := 0; r < n; r++ {
		for f := 0; f < n; f++ {
			sq := (n-1-r)*n + f
			sx, sy := x+f*cw, y+r*ch
			bg := l.style.a
			if (r+f)%2 == 1 {
				bg = l.style.b
			}
			if l.grey {
				bg = "line"
				if (r+f)%2 == 1 {
					bg = "bg"
				}
			}
			switch {
			case sq == l.cursor:
				bg = "cyanB"
			case sq == l.landing:
				bg = "cyanB"
			case l.path[sq]:
				bg = "greenB"
			case lit[sq]:
				bg = "yellowD"
			case l.last[sq]:
				bg = "line"
			}
			c.Shade(sx, sy, cw, ch, bg)
			stack := bd.Stacks[sq]
			if len(stack) == 0 && !l.grey {
				decorate(c, l.style.decor, sx, sy, cw, ch, r, f, bg)
			}
			drawStack(c, stack, sx, sy, cw, ch, l.set, l.ghost)
		}
	}
	if l.coords {
		for f := 0; f < n; f++ {
			c.Text(x+f*cw+cw/2, y+n*ch, string(rune('a'+f)), "dim", "", false)
		}
	}
}

// decorate dresses an empty square in the board's style.
func decorate(c *arcade.Canvas, decor string, x, y, cw, ch, r, f int, bg string) {
	switch {
	case decor == "rake" && cw > 2:
		for i := 1; i < cw-1; i++ {
			c.Text(x+i, y+ch-1, "~", "orangeD", bg, false)
		}
	case decor == "petals" && (r*3+f)%4 == 0:
		c.Text(x+cw-2, y, "✿", "pink", bg, false)
	case decor == "lanterns" && (r+f*2)%5 == 0:
		c.Text(x+cw/2, y+ch/2, "◆", "orange", bg, false)
	case decor == "vines" && (r*2+f)%3 == 0:
		c.Text(x+1, y, "❦", "green", bg, false)
	}
}

// drawStack draws a stack side on in a cw x ch square: layers from the
// bottom up, the top stone's shape on top, and the height in the corner.
func drawStack(c *arcade.Canvas, stack []engine.Piece, x, y, cw, ch int, s *stoneSet, ghost bool) {
	if len(stack) == 0 {
		return
	}
	ph := 2 * ch
	small := ch < 2
	room := ph - 3 // pixels for layers, leaving room for a wall or dome on top
	if small {
		room = ph - 1
	}
	room = max(1, room)
	wide := max(3, cw-2)
	x0 := x + (cw-wide)/2
	shown := stack[max(0, len(stack)-room):]
	py := 2*y + ph - 1
	for k, p := range shown {
		face, edge := s.colours(p.Color, ghost)
		top := k == len(shown)-1
		switch {
		case top && p.Kind == engine.Wall: // standing, two wide, up to three tall
			wx := x0 + wide/2 - 1
			for j := 0; j < 3 && py-j >= 2*y; j++ {
				c.Px(wx, py-j, edge)
				c.Px(wx+1, py-j, face)
			}
		case top && p.Kind == engine.Cap: // a dome
			cx := x0 + wide/2 - 1
			for i := -1; i < 3; i++ {
				c.Px(cx+i, py, edge)
			}
			if py-1 >= 2*y {
				c.Px(cx, py-1, face)
				c.Px(cx+1, py-1, face)
			}
		default:
			for i := 0; i < wide; i++ {
				role := face
				if i == 0 || i == wide-1 {
					role = edge
				}
				c.Px(x0+i, py, role)
			}
			py--
		}
	}
	if len(stack) > 1 && !small {
		role := "dim"
		if len(stack) > room {
			role = "yellow" // more than it can show
		}
		c.Text(x, y, strconv.Itoa(len(stack)), role, "", true)
	}
}

// drawStones draws a set's flat, wall and capstone for both sides, white
// above black, centred in w x h.
func drawStones(c *arcade.Canvas, s *stoneSet, x, y, w, h int, ghost bool) {
	x += max(0, (w-18)/2)
	rows := 2
	if h < 4 {
		rows = 1
	}
	for side := 0; side < rows; side++ {
		for k, kind := range []int8{engine.Flat, engine.Wall, engine.Cap} {
			drawStack(c, []engine.Piece{{Color: int8(side), Kind: kind}}, x+k*6, y+side*2, 6, 2, s, ghost)
		}
	}
}

// samplePosition is a game in progress, for previews.
func samplePosition() *engine.Board {
	b, _ := engine.New(5, 0)
	put := func(sq int, pieces ...engine.Piece) { b.Stacks[sq] = append(b.Stacks[sq], pieces...) }
	w, k := int8(engine.White), int8(engine.Black)
	put(16, engine.Piece{Color: w}, engine.Piece{Color: k}, engine.Piece{Color: w})
	put(17, engine.Piece{Color: k, Kind: engine.Wall})
	put(18, engine.Piece{Color: w, Kind: engine.Cap})
	put(13, engine.Piece{Color: k})
	put(12, engine.Piece{Color: w}, engine.Piece{Color: w}, engine.Piece{Color: k}, engine.Piece{Color: w})
	put(11, engine.Piece{Color: k})
	put(8, engine.Piece{Color: w, Kind: engine.Wall})
	put(7, engine.Piece{Color: k, Kind: engine.Cap})
	put(6, engine.Piece{Color: w})
	return b
}

// preview draws an unlockable for the arcade: a stone set as its stones,
// or a sample game in it when there's room; a board as a sample game on it.
// Only the locked thing is greyed out.
func preview(c *arcade.Canvas, id string, sel map[string]string, x, y, w, h int, ghost bool) {
	s, st := set(sel[kindStones]), style(sel[kindBoard])
	isBoard := len(id) > len(boardPrefix) && id[:len(boardPrefix)] == boardPrefix
	if isBoard {
		st = style(id)
	} else {
		s = set(id)
	}
	if h >= 8 || isBoard {
		drawBoard(c, x, y, w, h, samplePosition(), look{set: s, style: st, cursor: -1, landing: -1,
			ghost: ghost && !isBoard, grey: ghost && isBoard})
		return
	}
	drawStones(c, s, x, y, w, h, ghost)
}

// setupPreview is the NEW GAME screen's picture: an empty board of the
// chosen size with a stone of each kind.
func setupPreview(c *arcade.Canvas, options map[string]string, x, y, w, h int) {
	g := New(options)
	b := g.B
	n := b.N
	b.Stacks[0] = []engine.Piece{{Color: engine.White}}
	b.Stacks[n-1] = []engine.Piece{{Color: engine.Black, Kind: engine.Wall}}
	b.Stacks[n*n-1] = []engine.Piece{{Color: engine.White, Kind: engine.Cap}}
	drawBoard(c, x, y, w, h, b, look{set: &stoneSets[0], style: &boardStyles[0], cursor: -1, landing: -1, coords: true})
}

// roadOf is a winning road for color, edge to edge, or nil.
func roadOf(b *engine.Board, color int) []int {
	n := b.N
	ok := func(sq int) bool {
		p, there := b.Top(sq)
		return there && int(p.Color) == color && p.Kind != engine.Wall
	}
	for _, vertical := range []bool{false, true} {
		prev := map[int]int{}
		var queue []int
		for i := 0; i < n; i++ {
			sq := i * n
			if vertical {
				sq = i
			}
			if ok(sq) {
				prev[sq], queue = -1, append(queue, sq)
			}
		}
		for len(queue) > 0 {
			sq := queue[0]
			queue = queue[1:]
			if (!vertical && sq%n == n-1) || (vertical && sq/n == n-1) {
				var path []int
				for at := sq; at != -1; at = prev[at] {
					path = append([]int{at}, path...)
				}
				return path
			}
			for dir := 0; dir < 4; dir++ {
				if nx := step(b, sq, dir); nx >= 0 && ok(nx) {
					if _, seen := prev[nx]; !seen {
						prev[nx], queue = sq, append(queue, nx)
					}
				}
			}
		}
	}
	return nil
}

// ── Callouts ───────────────────────────────────────────────────────────────

const calloutLasts = 1400 * time.Millisecond

// drawCallout writes text in the logo font, centred on (cx, row), on a
// plain backing so it reads over the board.
func drawCallout(c *arcade.Canvas, text string, cx, row int) {
	w := arcade.LogoWidth(text, 1)
	x := cx - w/2
	for j := row - 1; j < row+arcade.LogoHeight(1)+1; j++ { // clear what's under it, pixels too
		c.Fill(x-1, j, w+3, " ", "fg", "bg")
	}
	c.Logo(text, x, 2*row, 1, []string{"yellow", "yellow", "yellow", "orange", "orange", "orange", "orange"})
}

// ── Attract mode ───────────────────────────────────────────────────────────

// The computer plays itself on a 5x5: two seeded opening moves, then the
// engine's normal player (which always picks the same move in the same
// position), a move every demoStep ticks, a pause on the end, then the next
// game in the next set on the next board. Positions are worked out as
// they're first needed and kept, so every frame of a game is the same
// picture whoever asks.
const (
	demoStep  = 3
	demoPlies = 40
	demoPause = 8
)

type demoFrame struct {
	b    *engine.Board
	last map[int]bool
}

var demo struct {
	n      int
	frames []demoFrame
	done   bool
}

// demoFrames returns game n's positions up to ply (fewer when it's over).
func demoFrames(n, ply int) ([]demoFrame, bool) {
	if demo.frames == nil || demo.n != n {
		b, _ := engine.New(5, 0)
		demo.n, demo.frames, demo.done = n, []demoFrame{{b: b}}, false
	}
	for !demo.done && len(demo.frames) <= ply {
		b := demo.frames[len(demo.frames)-1].b
		var m engine.Move
		if b.Ply < 2 {
			moves := b.Moves()
			m = moves[(n*7+b.Ply*11)%len(moves)]
		} else {
			m = engine.Best(b, 2)
		}
		nb := b.Apply(m)
		last := map[int]bool{m.Sq: true}
		at := m.Sq
		for range m.Drops {
			at = step(b, at, m.Dir)
			last[at] = true
		}
		demo.frames = append(demo.frames, demoFrame{b: nb, last: last})
		if r := nb.Result(); r.Over || nb.Ply >= demoPlies {
			demo.done = true
		}
	}
	return demo.frames, demo.done
}

func attract(c *arcade.Canvas, x, y, w, h, frame int) {
	step := frame / demoStep
	n := int(time.Now().Unix() / 3600)
	frames, done := demoFrames(n, step)
	for done && step >= len(frames)+demoPause && n < 1<<30 {
		step -= len(frames) + demoPause
		n++
		frames, done = demoFrames(n, step)
	}
	f := frames[min(step, len(frames)-1)]
	s := &stoneSets[n%len(stoneSets)]
	st := &boardStyles[n%len(boardStyles)]
	l := look{set: s, style: st, cursor: -1, landing: -1, last: f.last}
	r := f.b.Result()
	if r.Over && r.Winner >= 0 {
		l.road = roadOf(f.b, r.Winner)
		l.lit = frame % 40 / 2
	}
	bw := 42
	drawBoard(c, x, y+1, bw, h-2, f.b, l)
	px := x + bw + 4
	c.Text(px, y+1, "NOW SHOWING", "pink", "", true)
	c.Text(px, y+3, s.Name, s.Tier.Role(), "", true)
	c.Text(px, y+4, "on "+st.Name, "dim", "", false)
	if s.Tier != arcade.Starter {
		c.Text(px, y+6, s.Tier.Stars()+" "+s.Tier.Name(), s.Tier.Role(), "", false)
	}
	c.Text(px, y+8, "MOVE "+strconv.Itoa((f.b.Ply+1)/2), "dim", "", false)
	if r.Over && len(l.road) > 0 {
		c.Text(px, y+9, "★ ROAD!", "yellow", "", true)
	} else if r.Over {
		c.Text(px, y+9, "★ FLATS!", "yellow", "", true)
	}
}
