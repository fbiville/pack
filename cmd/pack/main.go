package main

import (
	"fmt"
	"github.com/buildpack/pack/style"
	"github.com/fatih/color"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/buildpack/pack"
	"github.com/buildpack/pack/config"
	"github.com/buildpack/pack/docker"
	"github.com/buildpack/pack/fs"
	"github.com/buildpack/pack/image"
)

var Version = "0.0.0"
var noTimestamps, verbose bool
var logger *pack.Logger

func main() {
	rootCmd := &cobra.Command{
		Use: "pack",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger = pack.NewLogger(os.Stdout, os.Stderr, verbose, noTimestamps)
		},
	}
	rootCmd.PersistentFlags().BoolVar(&color.NoColor, "no-color", false, "Disable color output")
	rootCmd.PersistentFlags().BoolVar(&noTimestamps, "no-timestamps", false, "Disable timestamps in output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Show verbose output")
	rootCmd.Flags().BoolP("help", "h", false, "Help for pack")
	for _, f := range []func() *cobra.Command{
		buildCommand,
		runCommand,
		rebaseCommand,
		createBuilderCommand,
		addStackCommand,
		updateStackCommand,
		deleteStackCommand,
		setDefaultStackCommand,
		versionCommand,
	} {
		rootCmd.AddCommand(f())
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildCommand() *cobra.Command {
	var buildFlags pack.BuildFlags
	cmd := &cobra.Command{
		Use:   "build <image-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Generate app image from source code",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			buildFlags.RepoName = args[0]
			bf, err := pack.DefaultBuildFactory()
			if err != nil {
				return err
			}
			b, err := bf.BuildConfigFromFlags(&buildFlags)
			if err != nil {
				return err
			}
			return b.Run()
		},
	}
	cmd.Flags().StringVarP(&buildFlags.AppDir, "path", "p", "current working directory", "Path to app dir")
	cmd.Flags().StringVar(&buildFlags.Builder, "builder", "packs/samples", "Builder")
	cmd.Flags().StringVar(&buildFlags.RunImage, "run-image", "default stack run image", "Run image")
	cmd.Flags().BoolVar(&buildFlags.Publish, "publish", false, "Publish to registry")
	cmd.Flags().BoolVar(&buildFlags.NoPull, "no-pull", false, "Skip pulling images before use")
	cmd.Flags().StringSliceVar(&buildFlags.Buildpacks, "buildpack", []string{}, "Buildpack ID, path to directory, or path/URL to .tgz file"+multiValueHelp("buildpack"))
	cmd.Flags().BoolP("help", "h", false, "Help for build")
	return cmd
}

func runCommand() *cobra.Command {
	wd, _ := os.Getwd()

	var runFlags pack.RunFlags
	cmd := &cobra.Command{
		Use:   "run",
		Args:  cobra.NoArgs,
		Short: "Build and run app image (recommended for development only)",
		RunE: func(cmd *cobra.Command, args []string) error {
			bf, err := pack.DefaultBuildFactory()
			if err != nil {
				return err
			}
			r, err := bf.RunConfigFromFlags(&runFlags)
			if err != nil {
				return err
			}
			cmd.SilenceUsage = true
			return r.Run(makeStopChannelForSignals)
		},
	}
	cmd.Flags().StringVarP(&runFlags.AppDir, "path", "p", wd, "Path to app directory")
	cmd.Flags().StringVar(&runFlags.Builder, "builder", "packs/samples", "Builder")
	cmd.Flags().StringVar(&runFlags.RunImage, "run-image", "default stack run image", "Run image")
	cmd.Flags().StringSliceVar(&runFlags.Ports, "port", nil, "Comma separated ports to publish (defaults to ports exposed by container)")
	cmd.Flags().BoolP("help", "h", false, "Help for run")
	return cmd
}

func rebaseCommand() *cobra.Command {
	var flags pack.RebaseFlags
	cmd := &cobra.Command{
		Use:   "rebase <image-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Rebase app image with latest run image",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			flags.RepoName = args[0]

			imageFactory, err := image.DefaultFactory()
			if err != nil {
				return err
			}
			cfg, err := config.NewDefault()
			if err != nil {
				return err
			}
			factory := pack.RebaseFactory{
				Log:          log.New(os.Stdout, "", log.LstdFlags),
				Config:       cfg,
				ImageFactory: imageFactory,
			}
			rebaseConfig, err := factory.RebaseConfigFromFlags(flags)
			if err != nil {
				return err
			}
			return factory.Rebase(rebaseConfig)
		},
	}
	cmd.Flags().BoolVar(&flags.Publish, "publish", false, "Publish to registry")
	cmd.Flags().BoolVar(&flags.NoPull, "no-pull", false, "Skip pulling images before use")
	cmd.Flags().BoolP("help", "h", false, "Help for rebase")
	return cmd
}

func createBuilderCommand() *cobra.Command {
	flags := pack.CreateBuilderFlags{}
	cmd := &cobra.Command{
		Use:   "create-builder <image-name> --builder-config <builder-toml-path>",
		Args:  cobra.ExactArgs(1),
		Short: "Create builder image",
		RunE: runE(func(cmd *cobra.Command, args []string) error {
			flags.RepoName = args[0]

			dockerClient, err := docker.New()
			if err != nil {
				return err
			}
			cfg, err := config.NewDefault()
			if err != nil {
				return err
			}
			builderFactory := pack.BuilderFactory{
				FS:     &fs.FS{},
				Logger: logger,
				Docker: dockerClient,
				Config: cfg,
				Images: &image.Client{},
			}
			builderConfig, err := builderFactory.BuilderConfigFromFlags(flags)
			if err != nil {
				return err
			}
			return builderFactory.Create(builderConfig)
		}),
	}
	cmd.Flags().BoolVar(&flags.NoPull, "no-pull", false, "Skip pulling stack image before use")
	cmd.Flags().StringVarP(&flags.BuilderTomlPath, "builder-config", "b", "", "Path to builder TOML file (required)")
	cmd.MarkFlagRequired("builder-config")
	cmd.Flags().StringVarP(&flags.StackID, "stack", "s", "", "Stack ID (defaults to stack configured by 'set-default-stack')")
	cmd.Flags().BoolVar(&flags.Publish, "publish", false, "Publish to registry")
	addHelpFlag(cmd, "create-builder")
	return cmd
}

func addStackCommand() *cobra.Command {
	flags := struct {
		BuildImage string
		RunImages  []string
	}{}
	cmd := &cobra.Command{
		Use:   "add-stack <stack-id> --build-image <build-image-name> --run-image <run-image-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Add stack to list of available stacks",
		RunE: runE(func(cmd *cobra.Command, args []string) error {
			cfg, err := config.NewDefault()
			if err != nil {
				return err
			}
			if err := cfg.Add(config.Stack{
				ID:         args[0],
				BuildImage: flags.BuildImage,
				RunImages:  flags.RunImages,
			}); err != nil {
				return err
			}
			logger.Info("Stack %s added", style.Emphasized(args[0]))
			return nil
		}),
	}
	cmd.Flags().StringVarP(&flags.BuildImage, "build-image", "b", "", "Build image to associate with stack (required)")
	cmd.MarkFlagRequired("build-image")
	cmd.Flags().StringSliceVarP(&flags.RunImages, "run-image", "r", nil, "Run image to associate with stack (required)"+multiValueHelp("run image"))
	cmd.MarkFlagRequired("run-image")
	addHelpFlag(cmd, "add-stack")
	return cmd
}

func updateStackCommand() *cobra.Command {
	flags := struct {
		BuildImage string
		RunImages  []string
	}{}
	cmd := &cobra.Command{
		Use:   "update-stack <stack-id> --build-image <build-image-name> --run-image <run-image-name>",
		Args:  cobra.ExactArgs(1),
		Short: "Update stack build and run images",
		RunE: runE(func(cmd *cobra.Command, args []string) error {
			cfg, err := config.New(filepath.Join(os.Getenv("HOME"), ".pack"))
			if err != nil {
				return err
			}
			if err := cfg.Update(args[0], config.Stack{
				BuildImage: flags.BuildImage,
				RunImages:  flags.RunImages,
			}); err != nil {
				return err
			}
			logger.Info("Stack %s updated", style.Emphasized(args[0]))
			return nil
		}),
	}
	cmd.Flags().StringVarP(&flags.BuildImage, "build-image", "b", "", "Build image to associate with stack (required)")
	cmd.MarkFlagRequired("build-image")
	cmd.Flags().StringSliceVarP(&flags.RunImages, "run-image", "r", nil, "Run image to associate with stack (required)"+multiValueHelp("run image"))
	cmd.MarkFlagRequired("run-image")
	addHelpFlag(cmd, "update-stack")
	return cmd
}

func deleteStackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete-stack <stack-id>",
		Args:  cobra.ExactArgs(1),
		Short: "Delete stack from list of available stacks",
		RunE: runE(func(cmd *cobra.Command, args []string) error {
			cfg, err := config.New(filepath.Join(os.Getenv("HOME"), ".pack"))
			if err != nil {
				return err
			}
			if err := cfg.Delete(args[0]); err != nil {
				return err
			}
			logger.Info("Stack %s deleted", style.Emphasized(args[0]))
			return nil
		}),
	}
	addHelpFlag(cmd, "delete-stack")
	return cmd
}

func setDefaultStackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-default-stack <stack-id>",
		Args:  cobra.ExactArgs(1),
		Short: "Set default stack used by other commands",
		RunE: runE(func(cmd *cobra.Command, args []string) error {
			cfg, err := config.New(filepath.Join(os.Getenv("HOME"), ".pack"))
			if err != nil {
				return err
			}
			err = cfg.SetDefaultStack(args[0])
			if err != nil {
				return err
			}
			logger.Info("Stack %s is now the default stack\n", style.Emphasized(args[0]))
			return nil
		}),
	}
	addHelpFlag(cmd, "set-default-stack")
	return cmd
}

func versionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Args:  cobra.NoArgs,
		Short: "Show current pack version",
		RunE: runE(func(cmd *cobra.Command, args []string) error {
			logger.Info(strings.TrimSpace(Version))
			return nil
		}),
	}
	addHelpFlag(cmd, "version")
	return cmd
}

func makeStopChannelForSignals() <-chan struct{} {
	sigsCh := make(chan os.Signal, 1)
	stopCh := make(chan struct{}, 1)
	signal.Notify(sigsCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		// convert chan os.Signal to chan struct{}
		for {
			<-sigsCh
			stopCh <- struct{}{}
		}
	}()
	return stopCh
}

func addHelpFlag(cmd *cobra.Command, commandName string) {
	cmd.Flags().BoolP("help", "h", false, "Help for "+commandName)
}

func runE(f func(cmd *cobra.Command, args []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cmd.SilenceErrors = true
		cmd.SilenceUsage = true
		err := f(cmd, args)
		if err != nil {
			logger.Error(err.Error())
			return err
		}
		return nil
	}
}

func multiValueHelp(name string) string {
	return fmt.Sprintf(",\nrepeat for each %s in order,\nor supply once with comma-separated list", name)
}
