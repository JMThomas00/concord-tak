# Tak

Tak, the abstract strategy game, for the terminal and for
[Concord](https://github.com/JMThomas00/Concord) channels, from the same
program.

## Play it on your own computer

**Download:** from the [Releases](https://github.com/JMThomas00/concord-tak/releases)
page, get the zip for your system (`concord-tak_windows_amd64.zip`,
`concord-tak_darwin_arm64.zip` for Apple silicon, `concord-tak_linux_amd64.zip`, ...),
unzip it, and run the program inside from a terminal:

```sh
./concord-tak          # Windows: .\concord-tak.exe
```

On macOS, if it's blocked as being from an unidentified developer, run
`xattr -d com.apple.quarantine concord-tak` once. **Or with Go installed:**
`go install github.com/JMThomas00/concord-tak@latest`, then run `concord-tak`.

It starts with a menu: two players on one keyboard, against the computer at
three levels, or over the network. For a network game, one player hosts and is
shown their address and a 6-character code; the other chooses join and types
both. Pick the board with `-size 3` to `-size 8` (default 5), and give Black extra points in a flat count with `-komi 2`.

## Play it on a Concord server

You need to be the server owner, or have the **Manage Plugins** permission.

1. In Concord, open **Server Settings → Plugins** and press **I** (install).
2. Type `JMThomas00/concord-tak` and press Enter. Concord downloads the latest
   release for the server's own system, verifies it, and starts it: no
   restart, no files to edit.
3. Open **Server Settings → Channels**, create a channel, and choose
   **Tak Table** as its type. Its options:
   - **Seating**: *seats* (one board; sit down with M, everyone else
     watches), *challenge* (a lobby where members challenge each other), or
     *private* (your own games with opponents you pick).
   - **Allow spectators**, **Computer opponent**, and **Computer strength**
     (easy, normal or hard).
   - **Board size** (3×3 to 8×8) and **Komi**.
4. Select the channel and press **Tab** (or click the board) so your keys go
   to the game. **M** opens the table menu: sit down, play the computer,
   resign, rematch. **Esc** gives the keyboard back to Concord (once there's
   nothing to cancel), and **Tab** moves on to the member list.

To update later: select it in **Server Settings → Plugins**, press **U**, then
Enter. A failed update rolls back by itself.

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
- **M** opens the table menu: resign, rematch, and so on. (playing standalone, Tab does too)

## Layout

- `engine/`: the rules, PTN, and the computer player. Move generation is
  checked against the known move counts (perft) for 5×5.
- `game/`: connects the engine to the Concord SDK's table kit, and the board you play on.
- `release.go`: `go run release.go` builds the release zips Concord installs.

Tag a version (`git tag v0.1.0 && git push --tags`) and the workflow publishes them.
