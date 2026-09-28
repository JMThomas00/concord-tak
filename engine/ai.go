package engine

import (
	"math/rand"
	"sort"
	"time"
)

// Best picks a move for the side to move:
//
//	1 (easy)   -- one move ahead, with some randomness
//	2 (normal) -- two moves ahead (sees your threats)
//	3 (hard)   -- as deep as it gets in about 2.5 seconds
func Best(b *Board, level int) Move {
	moves := b.Moves()
	// A move that wins on the spot is always right.
	for _, m := range moves {
		if r := b.Apply(m).Result(); r.Over && r.Winner == b.ToMove {
			return m
		}
	}
	switch level {
	case 1:
		scored := (&search{}).root(b, moves, 1)
		best := scored[0].score
		var ok []Move
		for _, s := range scored {
			if s.score >= best-60 {
				ok = append(ok, s.move)
			}
		}
		return ok[rand.Intn(len(ok))]
	case 2:
		return (&search{}).root(b, moves, 2)[0].move
	default:
		deadline := time.Now().Add(2500 * time.Millisecond)
		best := (&search{}).root(b, moves, 2)
		for depth := 3; depth <= 6 && time.Now().Before(deadline); depth++ {
			s := &search{deadline: deadline}
			scored := s.root(b, orderFrom(best), depth)
			if s.timeout {
				break
			}
			best = scored
		}
		return best[0].move
	}
}

const winScore = 1000000

type scored struct {
	move  Move
	score int
}

func orderFrom(s []scored) []Move {
	out := make([]Move, len(s))
	for i := range s {
		out[i] = s[i].move
	}
	return out
}

type search struct {
	deadline time.Time
	nodes    int
	timeout  bool
}

func (s *search) root(b *Board, moves []Move, depth int) []scored {
	out := make([]scored, len(moves))
	alpha := -winScore - 1
	for i, m := range moves {
		score := -s.negamax(b.Apply(m), depth-1, -winScore-1, -alpha)
		out[i] = scored{m, score}
		if score > alpha {
			alpha = score
		}
		if s.timeout {
			break
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].score > out[j].score })
	return out
}

// negamax scores b for the side to move.
func (s *search) negamax(b *Board, depth, alpha, beta int) int {
	s.nodes++
	if s.nodes&255 == 0 && !s.deadline.IsZero() && time.Now().After(s.deadline) {
		s.timeout = true
	}
	if s.timeout {
		return 0
	}
	if r := b.Result(); r.Over {
		switch r.Winner {
		case -1:
			return 0
		case b.ToMove:
			return winScore + depth
		default:
			return -winScore - depth // losing sooner is worse
		}
	}
	if depth <= 0 {
		return evaluate(b)
	}
	moves := b.Moves()
	// Try placements of flats first: they're the moves that most often matter.
	sort.SliceStable(moves, func(i, j int) bool { return rank(moves[i]) < rank(moves[j]) })
	for _, m := range moves {
		score := -s.negamax(b.Apply(m), depth-1, -beta, -alpha)
		if score > alpha {
			alpha = score
		}
		if alpha >= beta {
			break
		}
	}
	return alpha
}

func rank(m Move) int {
	switch {
	case !m.IsSlide() && m.Kind == Flat:
		return 0
	case m.IsSlide():
		return 1
	case m.Kind == Cap:
		return 2
	}
	return 3
}

// evaluate scores a position for the side to move.
func evaluate(b *Board) int {
	var score [2]int
	for sq, st := range b.Stacks {
		if len(st) == 0 {
			continue
		}
		top := st[len(st)-1]
		c := top.Color
		switch top.Kind {
		case Flat:
			score[c] += 100
		case Wall:
			score[c] += 40
		case Cap:
			score[c] += 70
		}
		// Stones under your control: your own are reserves, theirs are captives.
		for _, p := range st[:len(st)-1] {
			if p.Color == c {
				score[c] += 12
			} else {
				score[c] += 6
			}
		}
		// A little for the middle of the board.
		f, r := sq%b.N, sq/b.N
		if f > 0 && f < b.N-1 && r > 0 && r < b.N-1 {
			score[c] += 8
		}
	}
	for c := 0; c < 2; c++ {
		score[c] += roadPotential(b, c)
	}
	// Flats decide the game if the board fills or supplies run out; weigh
	// them more as the game nears its end.
	left := b.Stones[0] + b.Stones[1]
	if left < b.N*2 {
		f := b.Flats()
		score[Black] += b.Komi2 * 50
		score[White] += f[White] * 40
		score[Black] += f[Black] * 40
	}
	me := b.ToMove
	return score[me] - score[1-me]
}

// roadPotential rewards connected groups of flats and capstones by how
// far they stretch across the board (a group that spans it is a road).
func roadPotential(b *Board, color int) int {
	n := b.N
	seen := make([]bool, n*n)
	total := 0
	for start := range b.Stacks {
		if seen[start] || !roadStone(b, start, color) {
			continue
		}
		minF, maxF, minR, maxR := n, -1, n, -1
		queue := []int{start}
		seen[start] = true
		for len(queue) > 0 {
			sq := queue[0]
			queue = queue[1:]
			f, r := sq%n, sq/n
			minF, maxF, minR, maxR = min(minF, f), max(maxF, f), min(minR, r), max(maxR, r)
			for dir := 0; dir < 4; dir++ {
				if nx := b.step(sq, dir); nx >= 0 && !seen[nx] && roadStone(b, nx, color) {
					seen[nx] = true
					queue = append(queue, nx)
				}
			}
		}
		span := max(maxF-minF, maxR-minR) + 1
		total += span * span * 12
	}
	return total
}

func roadStone(b *Board, sq, color int) bool {
	top, ok := b.Top(sq)
	return ok && int(top.Color) == color && top.Kind != Wall
}
