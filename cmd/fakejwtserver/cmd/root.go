package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/stackitcloud/fake-jwt-server/server"
)

var (
	audience string
	issuer   string
	subject  string
	port     int
	id       string
)

var rootCmd = &cobra.Command{
	Use:   "fake jwt server",
	Short: "A fake JWT server",
	Long:  `A fake JWT server that can be used to generate JWT tokens for testing purposes.`,
	Run: func(cmd *cobra.Command, args []string) {
		fakeJwtServer := server.NewFakeJwtServer()
		fakeJwtServer.WithAudience(audience)
		fakeJwtServer.WithIssuer(issuer)
		fakeJwtServer.WithSubject(subject)
		fakeJwtServer.WithID(id)
		fakeJwtServer.WithPort(port)

		err := fakeJwtServer.Serve()
		if err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&audience, "audience", "product", "The audience of the JWT token")
	rootCmd.PersistentFlags().StringVar(&issuer, "issuer", "product", "The issuer of the JWT token")
	rootCmd.PersistentFlags().StringVar(&subject, "subject", "test", "The subject of the JWT token")
	rootCmd.PersistentFlags().StringVar(&id, "id", "test", "The id of the JWT token")
	rootCmd.PersistentFlags().IntVar(&port, "port", 8008, "The port the server should listen on")
}

func initConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	// There is some issue, where the integration of Cobra with Viper will result in wrong values, therefore we are
	// setting the values from viper manually. The issue is, that with the standard integration, viper will see, that
	// Cobra parameters are set - even if the command line parameter was not used and the default value was set. But
	// when Viper notices that the value is set, it will not overwrite the default value with the environment variable.
	// Another possibility would be to not have any default values set for cobra command line parameters, but this would
	// break the automatic help output from the cli. The manual way here seems the best solution for now.
	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && viper.IsSet(f.Name) {
			if err := rootCmd.PersistentFlags().Set(f.Name, fmt.Sprint(viper.Get(f.Name))); err != nil {
				log.Fatalf("unable to set value for command line parameter: %v", err)
			}
		}
	})
}
