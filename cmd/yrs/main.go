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
			&cli.StringSliceFlag{
				Name:  "not",
				Usage: "Exclude posts containing these keywords (NOT search)",
			},
			&cli.StringFlag{
				Name:  "from",
				Usage: "Filter posts from a specific account (ID:xxx)",
			},
			&cli.StringFlag{
				Name:  "to",
				Usage: "Filter posts mentioning a specific account (@xxx)",
			},
			&cli.StringSliceFlag{
				Name:  "hashtag",
				Usage: "Filter posts containing hashtags (#xxx)",
			},
			&cli.StringFlag{
				Name:  "url",
				Usage: "Filter posts containing a URL/domain (URL:xxx)",
			},
			&cli.StringSliceFlag{
				Name:  "or",
				Usage: "OR keywords group (posts containing any of these)",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			query := cmd.Args().First()

			hasBuilderFlags := len(cmd.StringSlice("not")) > 0 ||
				cmd.String("from") != "" ||
				cmd.String("to") != "" ||
				len(cmd.StringSlice("hashtag")) > 0 ||
				cmd.String("url") != "" ||
				len(cmd.StringSlice("or")) > 0

			var queryStr string
			if hasBuilderFlags {
				qb := yrs.NewQuery()
				if query != "" {
					qb.And(query)
				}
				for _, kw := range cmd.StringSlice("not") {
					qb.Not(kw)
				}
				if from := cmd.String("from"); from != "" {
					qb.FromUser(from)
				}
				if to := cmd.String("to"); to != "" {
					qb.ToUser(to)
				}
				for _, tag := range cmd.StringSlice("hashtag") {
					qb.Hashtag(tag)
				}
				if u := cmd.String("url"); u != "" {
					qb.URL(u)
				}
				if orKeywords := cmd.StringSlice("or"); len(orKeywords) > 0 {
					qb.Or(orKeywords...)
				}
				var err error
				queryStr, err = qb.Build()
				if err != nil {
					return cli.Exit(fmt.Sprintf("Invalid query: %v", err), 1)
				}
			} else {
				queryStr = query
			}

			if queryStr == "" {
				return cli.Exit("Search query is required.", 1)
			}

			client := yrs.NewClient(
				yrs.WithRequestTimeout(cmd.Duration("timeout")),
			)

			limit := int(cmd.Int("limit"))
			var result *yrs.SearchResult
			var err error
			if limit > 0 {
				result, err = client.SearchWithLimit(ctx, queryStr, limit)
			} else {
				result, err = client.Search(ctx, queryStr)
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
