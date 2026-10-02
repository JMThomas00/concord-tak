package game

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/JMThomas00/Concord/sdk/table"
	"github.com/JMThomas00/Concord/sdk/wire"
	"github.com/JMThomas00/concord-tak/engine"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Board is what a player sees and plays on. Rank 1 is at the bottom for
// everyone, as Tak boards are usually drawn.
type Board struct {
	seat          *table.Seat
	file, rank    int // cursor
	width, height int
	err           string

	// Moving a stack: picked up from `from`, carrying `carry` stones, in
	// direction dir (-1 until the first arrow), dropping `drops`.
	carrying bool
	from     int
	carry    int
	dir      int
	drops    []int

	typing bool   // entering PTN after ':'
	input  string // what's been typed
}

func newBoard(s *table.Seat) *Board {
	n := s.Game().(*Game).B.N
	return &Board{seat: s, file: n / 2, rank: n / 2}
}

func (b *Board) game() *Game   { return b.seat.Game().(*Game) }
func (b *Board) n() int        { return b.game().B.N }
func (b *Board) cursor() int   { return b.rank*b.n() + b.file }
func (b *Board) Init() tea.Cmd { return nil }

// Arrow keys → engine directions (0 up, 1 down, 2 right, 3 left).
var arrowDir = map[string]int{"up": 0, "k": 0, "down": 1, "j": 1, "right": 2, "l": 2, "left": 3, "h": 3}

func (b *Board) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		b.width, b.height = msg.Width, msg.Height
	case table.ChangedMsg:
		b.err, b.carrying, b.typing = "", false, false
		n := b.n()
		b.file, b.rank = min(b.file, n-1), min(b.rank, n-1)
	case tea.KeyMsg:
		if b.typing {
			return b, b.typeKey(msg)
		}
		if b.carrying {
			b.carryKey(msg.String())
			return b, nil
		}
		key := msg.String()
		if d, ok := arrowDir[key]; ok {
			b.moveCursor(d)
			return b, nil
		}
		switch key {
		case "enter", " ", "f":
			b.act(engine.Flat)
		case "s":
			b.act(engine.Wall)
		case "c":
			b.act(engine.Cap)
		case ":":
			if b.seat.MyTurn() {
				b.typing, b.input, b.err = true, "", ""
			}
		case "q":
			return b, tea.Quit // hand the keyboard back (in Concord)
		}
	}
	return b, nil
}

// ClaimedKeys keeps Esc while there's something for it to cancel: a stack
// being carried, or a move being typed. Otherwise Esc is Concord's, to
// leave the pane.
func (b *Board) ClaimedKeys() []string {
	if b.carrying || b.typing {
		return []string{wire.PaneKeyEsc}
	}
	return nil
}

// Typing (table.Typer) sends every key to the board while a move is being
// typed, M included.
func (b *Board) Typing() bool { return b.typing }

func (b *Board) moveCursor(dir int) {
	n := b.n()
	switch dir {
	case 0:
		b.rank = min(n-1, b.rank+1)
	case 1:
		b.rank = max(0, b.rank-1)
	case 2:
		b.file = min(n-1, b.file+1)
	case 3:
		b.file = max(0, b.file-1)
	}
}

// act places a stone on an empty square, or picks up the stack you control.
func (b *Board) act(kind int8) {
	b.err = ""
	if !b.seat.MyTurn() {
		b.err = "not your turn"
		return
	}
	g := b.game()
	sq := b.cursor()
	top, occupied := g.B.Top(sq)
	if !occupied {
		b.play(engine.Move{Sq: sq, Kind: kind}.PTN(g.B))
		return
	}
	if kind != engine.Flat {
		b.err = "that square is taken"
		return
	}
	if g.B.Ply < 2 {
		b.err = "first, place your opponent's stone on an empty square"
		return
	}
	if int(top.Color) != g.B.ToMove {
		b.err = "you don't control that stack"
		return
	}
	height := len(g.B.Stacks[sq])
	b.carrying, b.from, b.carry, b.dir, b.drops = true, sq, min(height, g.B.N), -1, nil
}

// carryKey drives moving a stack: +/- or a digit sets how many to carry
// (before the first step), arrows step and drop one, space drops another
// on the current square, Enter drops the rest here, Esc puts it back.
func (b *Board) carryKey(key string) {
	g := b.game()
	height := len(g.B.Stacks[b.from])
	dropped := 0
	for _, d := range b.drops {
		dropped += d
	}
	switch {
	case key == "esc":
		b.carrying = false
		return
	case len(b.drops) == 0 && (key == "+" || key == "="):
		b.carry = min(b.carry+1, min(height, g.B.N))
		return
	case len(b.drops) == 0 && key == "-":
		b.carry = max(1, b.carry-1)
		return
	case len(b.drops) == 0 && len(key) == 1 && key[0] >= '1' && key[0] <= '8':
		n, _ := strconv.Atoi(key)
		if n <= min(height, g.B.N) {
			b.carry = n
		}
		return
	}
	if d, ok := arrowDir[key]; ok {
		if b.dir >= 0 && d != b.dir {
			b.err = "a stack moves in a straight line"
			return
		}
		b.dir = d
		b.drops = append(b.drops, 1)
		dropped++
	} else if (key == " " || key == "space") && len(b.drops) > 0 {
		b.drops[len(b.drops)-1]++
		dropped++
	} else if key == "enter" && len(b.drops) > 0 {
		b.drops[len(b.drops)-1] += b.carry - dropped
		dropped = b.carry
	} else {
		return
	}
	if dropped >= b.carry {
		move := engine.Move{Sq: b.from, Dir: b.dir, Drops: b.drops}.PTN(g.B)
		b.carrying = false
		b.play(move)
	}
}

func (b *Board) typeKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc:
		b.typing = false
	case tea.KeyEnter:
		b.typing = false
		b.play(strings.TrimSpace(b.input))
	case tea.KeyBackspace:
		if len(b.input) > 0 {
			b.input = b.input[:len(b.input)-1]
		}
	case tea.KeyRunes:
		b.input += string(msg.Runes)
	}
	return nil
}

func (b *Board) play(ptn string) {
	if err := b.seat.Play(ptn); err != nil {
		b.err = err.Error()
	}
}

// ── Drawing ─────────────────────────────────────────────────────────────────

func (b *Board) colorOf(c int8) lipgloss.TerminalColor {
	if c == engine.White {
		return b.seat.Color("foreground", lipgloss.Color("15"))
	}
	return b.seat.Color("purple", lipgloss.Color("13"))
}

func glyph(kind int8) string {
	switch kind {
	case engine.Wall:
		return "▌"
	case engine.Cap:
		return "▲"
	}
	return "●"
}

// path is where the carried stones would land, for highlighting.
func (b *Board) path() map[int]int {
	out := map[int]int{}
	if !b.carrying || b.dir < 0 {
		return out
	}
	g := b.game()
	at := b.from
	for _, d := range b.drops {
		at = step(g.B, at, b.dir)
		if at < 0 {
			break
		}
		out[at] = d
	}
	return out
}

func step(bd *engine.Board, sq, dir int) int {
	n := bd.N
	f, r := sq%n, sq/n
	switch dir {
	case 0:
		r++
	case 1:
		r--
	case 2:
		f++
	case 3:
		f--
	}
	if f < 0 || f >= n || r < 0 || r >= n {
		return -1
	}
	return r*n + f
}

func (b *Board) View() string {
	g := b.game()
	n := g.B.N
	// Square size: as big as fits, leaving room for labels and 3 status lines.
	cellW := 7
	for cellW > 3 && 2+n*cellW > b.width {
		cellW -= 2
	}
	cellH := 3
	for cellH > 1 && n*cellH+4 > b.height {
		cellH--
	}
	even := b.seat.Color("current_line", lipgloss.Color("237"))
	odd := b.seat.Color("selection", lipgloss.Color("239"))
	cur := b.seat.Color("cyan", lipgloss.Color("14"))
	hint := b.seat.Color("green", lipgloss.Color("10"))
	dim := lipgloss.NewStyle().Foreground(b.seat.Color("comment", lipgloss.Color("8")))
	path := b.path()

	var lines []string
	for r := n - 1; r >= 0; r-- {
		rows := make([]string, cellH)
		for i := range rows {
			label := "  "
			if i == cellH/2 {
				label = dim.Render(fmt.Sprintf("%d ", r+1))
			}
			rows[i] = label
		}
		for f := 0; f < n; f++ {
			sq := r*n + f
			bg := even
			if (r+f)%2 == 1 {
				bg = odd
			}
			style := lipgloss.NewStyle().Background(bg).Width(cellW).Align(lipgloss.Center)
			content := ""
			if top, ok := g.B.Top(sq); ok {
				content = glyph(top.Kind)
				if h := len(g.B.Stacks[sq]); h > 1 {
					content += strconv.Itoa(h)
				}
				style = style.Foreground(b.colorOf(top.Color)).Bold(true)
			}
			if d, ok := path[sq]; ok {
				content = strings.Repeat("·", d)
				style = style.Foreground(hint).Bold(true)
			}
			if b.carrying && sq == b.from {
				style = style.Underline(true)
			}
			if f == b.file && r == b.rank && b.seat.MyTurn() && !b.carrying {
				style = style.Background(cur)
			}
			for i := range rows {
				c := ""
				if i == cellH/2 {
					c = content
				}
				rows[i] += style.Render(c)
			}
		}
		lines = append(lines, rows...)
	}
	files := "  "
	for f := 0; f < n; f++ {
		files += lipgloss.NewStyle().Width(cellW).Align(lipgloss.Center).Render(string(rune('a' + f)))
	}
	lines = append(lines, dim.Render(files))

	// The stack under the cursor, bottom to top.
	sq := b.cursor()
	var stack strings.Builder
	stack.WriteString(g.B.Name(sq) + ": ")
	for _, p := range g.B.Stacks[sq] {
		stack.WriteString(lipgloss.NewStyle().Foreground(b.colorOf(p.Color)).Render(glyph(p.Kind)))
	}
	if len(g.B.Stacks[sq]) == 0 {
		stack.WriteString("empty")
	}
	supply := ""
	for c, name := range []string{"White", "Black"} {
		supply += fmt.Sprintf("   %s %d", name, g.B.Stones[c])
		if g.B.Caps[c] > 0 {
			supply += fmt.Sprintf("+%d▲", g.B.Caps[c])
		}
	}
	lines = append(lines, dim.Render(stack.String()+supply))

	status := ""
	switch {
	case b.typing:
		status = "move (PTN, e.g. c3 or 2c3>11): " + b.input + "█"
	case b.err != "":
		status = b.err
	case b.carrying:
		status = fmt.Sprintf("carrying %d — arrows move and drop one, space drops another, Enter drops the rest, Esc cancels", b.carry)
		if len(b.drops) == 0 {
			status = fmt.Sprintf("carrying %d (+/- to change) — arrow to move, Esc cancels", b.carry)
		}
	case b.seat.MyTurn() && g.B.Ply < 2:
		status = "place your opponent's first stone: Enter on an empty square"
	case b.seat.MyTurn():
		if top, ok := g.B.Top(sq); ok && int(top.Color) == g.B.ToMove {
			status = "your move: Enter picks up this stack · : to type a move"
		} else {
			status = "your move: Enter flat · s wall · c capstone · : type"
		}
	case g.Last != "":
		status = "last move " + g.Last
	}
	lines = append(lines, status)
	return strings.Join(lines, "\n")
}
