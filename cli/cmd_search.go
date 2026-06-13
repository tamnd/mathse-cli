package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/mathse-cli/mathse"
)

func (a *App) searchCmd() *cobra.Command {
	var sort string
	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search questions by title text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			n := a.effectiveLimit(10)
			opts := mathse.SearchOptions{
				Query:    args[0],
				Sort:     sort,
				PageSize: n,
			}
			a.progressf("searching for %q...", args[0])
			hits, err := a.client.Search(cmd.Context(), opts)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(hits, len(hits))
		},
	}
	cmd.Flags().StringVar(&sort, "sort", "votes", "sort order: votes|activity|newest")
	return cmd
}
