package game

import (
	"strings"

	"github.com/JMThomas00/Concord/sdk/arcade"
)

// The unlockables: stone sets and boards, traded for Pebbles at THE QUARRY.
// Each player sees the game in their own set on their own board.

const (
	kindStones = "stones"
	kindBoard  = "board"
	// Board ids are prefixed in the unlockables list so they can't clash
	// with a stone set's ("river" is both).
	boardPrefix = "board:"
)

// stoneSet colours each side's stones: a face and an edge (its shading).
type stoneSet struct {
	arcade.Unlockable
	white, black [2]string
}

func stones(id, name string, tier arcade.Tier, blurb string, white, black [2]string) stoneSet {
	return stoneSet{arcade.Unlockable{ID: id, Name: name, Tier: tier, Kind: kindStones, Blurb: blurb}, white, black}
}

var stoneSets = []stoneSet{
	stones("river", "RIVER STONES", arcade.Starter, "Smoothed by the stream.", [2]string{"fg", "dim"}, [2]string{"comment", "tire"}),
	stones("granite", "GRANITE & BASALT", arcade.Starter, "Heavy and honest.", [2]string{"dim", "fgD"}, [2]string{"tire", "line"}),
	stones("jade", "JADE & ONYX", arcade.Common, "Polished for the emperor.", [2]string{"green", "greenD"}, [2]string{"tire", "purpleB"}),
	stones("sand", "SANDSTONE & SLATE", arcade.Common, "Warm against cool.", [2]string{"orange", "orangeD"}, [2]string{"comment", "cyanB"}),
	stones("marble", "MARBLE", arcade.Common, "White veins, green veins.", [2]string{"fg", "cyanD"}, [2]string{"greenD", "tire"}),
	stones("wood", "WOODEN BLOCKS", arcade.Common, "From the toy box.", [2]string{"yellow", "orangeD"}, [2]string{"orangeD", "tire"}),
	stones("glass", "SEA GLASS", arcade.Rare, "Washed up on the beach.", [2]string{"cyan", "cyanD"}, [2]string{"green", "greenD"}),
	stones("coral", "CORAL & PEARL", arcade.Rare, "From the reef.", [2]string{"fg", "pinkB"}, [2]string{"pink", "pinkD"}),
	stones("gems", "GEMSTONES", arcade.Rare, "Ruby against sapphire.", [2]string{"red", "redD"}, [2]string{"cyan", "purpleB"}),
	stones("pancake", "PANCAKES & WAFFLES", arcade.Legendary, "Stack them high.", [2]string{"orange", "yellow"}, [2]string{"yellow", "orangeD"}),
	stones("books", "BOOKS", arcade.Legendary, "Read from the top.", [2]string{"red", "pink"}, [2]string{"cyan", "purple"}),
	stones("grapes", "GRAPE CRATES", arcade.Legendary, "From the vineyard.", [2]string{"green", "greenD"}, [2]string{"purple", "shadow"}),
}

// boardStyle is a board's two square colours and what decorates its empty
// squares.
type boardStyle struct {
	arcade.Unlockable
	a, b  string
	decor string // "rake", "petals", "lanterns", "vines" or ""
}

func board(id, name string, tier arcade.Tier, blurb, a, b, decor string) boardStyle {
	return boardStyle{arcade.Unlockable{ID: id, Name: name, Tier: tier, Kind: kindBoard, Blurb: blurb}, a, b, decor}
}

var boardStyles = []boardStyle{
	board("sand", "RAKED SAND", arcade.Starter, "Lines in the sand.", "yellowB", "orangeB", "rake"),
	board("slate", "SLATE", arcade.Starter, "Cool and grey.", "line", "bg", ""),
	board("moss", "MOSS", arcade.Common, "Soft underfoot.", "greenB", "line", ""),
	board("river", "RIVER BED", arcade.Common, "Stepping stones.", "cyanB", "purpleB", ""),
	board("blossom", "CHERRY BLOSSOM", arcade.Rare, "Petals on the path.", "pinkB", "line", "petals"),
	board("night", "LANTERN NIGHT", arcade.Rare, "Lit along the road.", "purpleB", "bg", "lanterns"),
	board("vineyard", "VINEYARD ROWS", arcade.Legendary, "Roads between the vines.", "greenB", "purpleB", "vines"),
}

// unlockables is every stone set and board, for the arcade.
func unlockables() []arcade.Unlockable {
	var out []arcade.Unlockable
	for _, s := range stoneSets {
		out = append(out, s.Unlockable)
	}
	for _, b := range boardStyles {
		u := b.Unlockable
		u.ID = boardPrefix + u.ID
		out = append(out, u)
	}
	return out
}

// set looks up a stone set (the first starter if id is unknown).
func set(id string) *stoneSet {
	for i := range stoneSets {
		if stoneSets[i].ID == id {
			return &stoneSets[i]
		}
	}
	return &stoneSets[0]
}

// style looks up a board (with or without its prefix).
func style(id string) *boardStyle {
	id = strings.TrimPrefix(id, boardPrefix)
	for i := range boardStyles {
		if boardStyles[i].ID == id {
			return &boardStyles[i]
		}
	}
	return &boardStyles[0]
}

// colours is a side's face and edge; ghost greys them out for a locked set.
func (s *stoneSet) colours(color int8, ghost bool) (face, edge string) {
	if ghost {
		return "comment", "ghost"
	}
	c := s.white
	if color == 1 {
		c = s.black
	}
	return c[0], c[1]
}
