package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/mathse-cli/mathse"
)

func (a *App) tagsCmd() *cobra.Command {
	var search string
	cmd := &cobra.Command{
		Use:   "tags",
		Short: "List or search tags on Math Stack Exchange",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(20)
			opts := mathse.TagsOptions{
				PageSize: n,
				Search:   search,
			}
			a.progressf("fetching tags...")
			tags, err := a.client.Tags(cmd.Context(), opts)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(tags, len(tags))
		},
	}
	cmd.Flags().StringVar(&search, "search", "", "filter tags by name fragment (e.g. calculus)")
	return cmd
}
