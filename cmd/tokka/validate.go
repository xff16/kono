package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/starwalkn/tokka"
)

var validateCmd = &cobra.Command{
	Use:          "validate",
	Short:        "Validates configuration file",
	Args:         cobra.NoArgs,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		err := runValidate()
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println("configuration file is valid, you can start the server")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}

func runValidate() error {
	if cfgPath == "" {
		cfgPath = os.Getenv("TOKKA_CONFIG")
	}
	if cfgPath == "" {
		cfgPath = "./tokka.json"
	}

	_, err := tokka.LoadConfig(cfgPath)
	if err != nil {
		return err
	}

	return nil
}
