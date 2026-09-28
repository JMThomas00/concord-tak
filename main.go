// Tak: the abstract strategy game, in a terminal or in Concord channels.
//
// Run it from a terminal to play two-player (one keyboard), against the
// computer, or over the network. Launched by a Concord server (which sets
// CONCORD_* variables), the same program hosts Tak tables in the server's
// channels.
package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/JMThomas00/Concord/sdk/plugin"
	"github.com/JMThomas00/Concord/sdk/table"
	"github.com/JMThomas00/concord-tak/game"
)

func main() {
	cfg, underConcord := plugin.ConfigFromEnv()
	if !underConcord {
		size := flag.Int("size", 5, "board size, 3 to 8")
		komi := flag.Float64("komi", 0, "points added to Black's flat count (e.g. 2 or 2.5)")
		flag.Parse()
		options := map[string]string{
			game.OptionSize: strconv.Itoa(*size),
			game.OptionKomi: strconv.FormatFloat(*komi, 'f', -1, 64),
		}
		if err := table.RunLocal(game.Rules, options); err != nil {
			log.Fatal(err)
		}
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if err := plugin.Run(ctx, cfg, table.New(game.Rules).Handler()); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}
