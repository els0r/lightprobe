package flags

import (
	"os"

	"github.com/spf13/cobra"
)

type Config struct {
	ID string
}

func Read() Config {
	var cfg = Config{}

	var rootCmd = &cobra.Command{
		Use:   "lightprobe",
		Short: "lighprobe is a conntrack-based network traffic meta-data capture process",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			// Do Stuff Here
		},
	}
	rootCmd.Flags().StringVarP(&cfg.ID, "id", "i", "", "ID bla")

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	return cfg
}
