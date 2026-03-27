package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	yrs "github.com/cnosuke/go-yahoo-realtime-search"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "yrs",
		Usage: "Search tweets via Yahoo! Japan Realtime Search",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "limit",
				Aliases: []string{"l"},
				Usage:   "Maximum number of tweets to return (0 = all)",
				Value:   0,
			},
			&cli.DurationFlag{
				Name:    "timeout",
				Aliases: []string{"t"},
				Usage:   "HTTP request timeout",
				Value:   30 * time.Second,
			},
			&cli.BoolFlag{
				Name:  "json",
				Usage: "Output results as JSON",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			query := cmd.Args().First()
			if query == "" {
				return cli.Exit("Search query is required.", 1)
			}

			client := yrs.NewClient(
				yrs.WithRequestTimeout(cmd.Duration("timeout")),
			)

			limit := int(cmd.Int("limit"))
			var result *yrs.SearchResult
			var err error
			if limit > 0 {
				result, err = client.SearchWithLimit(ctx, query, limit)
			} else {
				result, err = client.Search(ctx, query)
			}
			if err != nil {
				return cli.Exit(fmt.Sprintf("Search failed: %v", err), 1)
			}

			if cmd.Bool("json") {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}

			if len(result.Tweets) == 0 {
				fmt.Println("No results found.")
				return nil
			}

			for _, tw := range result.Tweets {
				fmt.Printf("@%s (%s) - %s\n", tw.ScreenName, tw.AuthorName, tw.CreatedAt.Format(time.RFC3339))
				fmt.Printf("  %s\n", tw.Text)
				fmt.Printf("  %s\n", tw.URL)
				fmt.Printf("  Reply:%d RT:%d Like:%d\n\n", tw.ReplyCount, tw.RTCount, tw.LikeCount)
			}
			fmt.Printf("Total: %d tweets\n", len(result.Tweets))
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
