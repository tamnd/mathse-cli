package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tamnd/mathse-cli/mathse"
)

func (a *App) answersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "answers <question-id>",
		Short: "Get answers for a question",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return codeError(exitUsage, fmt.Errorf("id must be a number: %w", err))
			}
			n := a.effectiveLimit(10)
			opts := mathse.AnswersOptions{PageSize: n}
			a.progressf("fetching answers for question %d...", id)
			answers, err := a.client.Answers(cmd.Context(), id, opts)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.renderOrEmpty(answers, len(answers))
		},
	}
}
