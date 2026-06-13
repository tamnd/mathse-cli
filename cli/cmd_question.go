package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func (a *App) questionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "question <id>",
		Short: "Get a single question by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.Atoi(args[0])
			if err != nil {
				return codeError(exitUsage, fmt.Errorf("id must be a number: %w", err))
			}
			a.progressf("fetching question %d...", id)
			q, err := a.client.Question(cmd.Context(), id)
			if err != nil {
				return mapFetchErr(err)
			}
			return a.render(q)
		},
	}
}
