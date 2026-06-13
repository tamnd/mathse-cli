package cli

import (
	"github.com/spf13/cobra"
	"github.com/tamnd/mathse-cli/mathse"
)

func (a *App) questionsCmd() *cobra.Command {
	var (
		sort string
		tag  string
	)
	cmd := &cobra.Command{
		Use:   "questions",
		Short: "List questions on Math Stack Exchange",
		RunE: func(cmd *cobra.Command, _ []string) error {
			n := a.effectiveLimit(10)
			opts := mathse.QuestionsOptions{
				Sort:     sort,
				Tag:      tag,
				PageSize: n,
			}
			a.progressf("fetching %d questions...", n)
			qs, err := a.client.Questions(cmd.Context(), opts)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(qs, len(qs))
		},
	}
	cmd.Flags().StringVar(&sort, "sort", "votes", "sort order: votes|activity|newest|unanswered")
	cmd.Flags().StringVar(&tag, "tag", "", "filter by tag (e.g. calculus)")
	return cmd
}
