# Tak

Tak, the abstract strategy game, for the terminal and for
[Concord](https://github.com/JMThomas00/Concord) channels, from the same
program.

- **In a terminal:** `go run .` offers two players on one keyboard, the computer at
  three levels, or a network game. For a network game, one player hosts and the other
  joins with the host's address and a 6-character code. `-size 3..8` picks the
  board (default 5) and `-komi 2` gives Black extra points in a flat count.
- **In Concord:** install it in **Server Settings → Plugins** (press **I** and
  type `JMThomas00/concord-tak`), then create a **Tak Table** channel. The
  channel's settings choose the board size, komi, and how people play:
  - **seats**: one board; sit down with Tab, and everyone else watches.
  - **challenge**: a lobby where members challenge each other.
  - **private**: your own games with opponents you pick.

## Rules

Standard rules. Players take turns to place a stone from their supply on an
empty square, or to move a stack they control (their stone is on top).
Each player's first move places one of the **opponent's** flat stones.

- A stone is placed **flat**, as a **wall** (standing stone) or, if you have one,
  as a **capstone**.
- A stack moves in a straight line: pick up to board-size stones off the top,
  and drop at least one on each square you pass. Nothing moves onto a wall or
  a capstone, except a capstone moving alone onto a wall, which flattens it.
- **Road win:** a connected line of your flats and capstones linking opposite
  edges. Walls don't count. If one move makes roads for both players, the
  player who moved wins.
- **Flat win:** when the board is full or either player runs out of stones,
  the player with more flats on top wins (plus komi for Black). Equal is a draw.

| Board | 3×3 | 4×4 | 5×5 | 6×6 | 7×7 | 8×8 |
|---|---|---|---|---|---|---|
| Stones | 10 | 15 | 21 | 30 | 40 | 50 |
| Capstones | 0 | 0 | 1 | 1 | 2 | 2 |

## Playing

- **Arrow keys** (or hjkl) move the cursor.
- **Enter** (or f) places a flat stone, **s** a wall, **c** a capstone.
- **Enter on your own stack** picks it up: **+/-** or a digit sets how many
  to carry, then each **arrow** step drops one stone, **Space** drops another
  on the same square, **Enter** drops the rest, **Esc** cancels.
- **:** types a move in PTN (`c3`, `Sd4`, `Cb2`, `3c3>12`).
- **Tab** opens the table menu: resign, rematch, and so on.

## Layout

- `engine/`: the rules, PTN, and the computer player. Move generation is
  checked against the known move counts (perft) for 5×5.
- `game/`: connects the engine to the Concord SDK's table kit, and the board you play on.
- `release.go`: `go run release.go` builds the release zips Concord installs.

Tag a version (`git tag v0.1.0 && git push --tags`) and the workflow publishes them.
