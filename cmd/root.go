/*
MIT License

Copyright (c) 2021 Fair Data Society

*/
package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	uds "github.com/asabya/go-ipc-uds"
	"github.com/fairDataSociety/stampkeeper/pkg"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	argSeparator = "$^~@@*"
)

var (
	cfgFile   string
	verbosity string

	logger     pkg.Logger
	keeper     *pkg.Keeper
	ctx        context.Context
	cancel     context.CancelFunc
	sockPath   = "stampkeeper.sock"
	socketPath string
	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "stampkeeper",
		Short: "Auto top up postage stamps",
		Long: `Stampkeeper can monitor multiple swarm postage stamps.
It will top them up and dilute stamps as required.`,
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	switch v := strings.ToLower(verbosity); v {
	case "0", "silent":
		logger = pkg.NewLogger(ioutil.Discard, 0)
	case "1", "error":
		logger = pkg.NewLogger(rootCmd.OutOrStdout(), logrus.ErrorLevel)
	case "2", "warn":
		logger = pkg.NewLogger(rootCmd.OutOrStdout(), logrus.WarnLevel)
	case "3", "info":
		logger = pkg.NewLogger(rootCmd.OutOrStdout(), logrus.InfoLevel)
	case "4", "debug":
		logger = pkg.NewLogger(rootCmd.OutOrStdout(), logrus.DebugLevel)
	case "5", "trace":
		logger = pkg.NewLogger(rootCmd.OutOrStdout(), logrus.TraceLevel)
	default:
		fmt.Println("unknown verbosity level ", v)
		os.Exit(1)
	}
	tmp := os.TempDir()
	socketPath = filepath.Join(tmp, sockPath)
	ctx, cancel = context.WithCancel(context.Background())
	if !uds.IsIPCListening(socketPath) {
		keeper = pkg.New(ctx, server, logger)
	}

	if len(os.Args) > 1 {
		if os.Args[1] != "start" && uds.IsIPCListening(socketPath) {
			opts := uds.Options{
				SocketPath: filepath.Join(tmp, sockPath),
			}
			r, w, c, err := uds.Dialer(opts)
			if err != nil {
				logger.Error(err)
				goto Execute
			}
			defer func() {
				err := c()
				if err != nil {
					logger.Error(err)
				}
			}()

			err = w(strings.Join(os.Args[1:], argSeparator))
			if err != nil {
				logger.Error(err)
				os.Exit(1)
			}
			v, err := r()
			if err != nil {
				logger.Error(err)
				os.Exit(1)

			}
			fmt.Println(v)
			return
		}
		if os.Args[1] == "start" {
			if uds.IsIPCListening(socketPath) {
				fmt.Println("Datahop daemon is already running")
				return
			}
			_, err := os.Stat(filepath.Join(tmp, sockPath))
			if !os.IsNotExist(err) {
				err := os.Remove(filepath.Join(tmp, sockPath))
				if err != nil {
					logger.Error(err)
					os.Exit(1)
				}
			}
			opts := uds.Options{
				SocketPath: filepath.Join(tmp, sockPath),
			}
			in, err := uds.Listener(context.Background(), opts)
			if err != nil {
				logger.Error(err)
				os.Exit(1)
			}
			go func() {
				for {
					client := <-in
					go func() {
						for {
							ip, err := client.Read()
							if err != nil {
								break
							}
							if len(ip) == 0 {
								break
							}
							commandStr := string(ip)
							var (
								childCmd *cobra.Command
								flags    []string
							)
							commandStr = strings.TrimSpace(commandStr)
							command := strings.Split(commandStr, argSeparator)

							if rootCmd.TraverseChildren {
								childCmd, flags, err = rootCmd.Traverse(command)
							} else {
								childCmd, flags, err = rootCmd.Find(command)
							}
							if err != nil {
								err = client.Write([]byte(err.Error()))
								if err != nil {
									logger.Error("Write error", err)
									client.Close()
								}
								break
							}
							childCmd.Flags().VisitAll(func(f *pflag.Flag) {
								err := f.Value.Set(f.DefValue)
								if err != nil {
									logger.Error("Unable to set flags ", childCmd.Name(), f.Name, err.Error())
								}
							})
							if err := childCmd.Flags().Parse(flags); err != nil {
								logger.Error("Unable to parse flags ", err.Error())
								err = client.Write([]byte(err.Error()))
								if err != nil {
									logger.Error("Write error", err)
									client.Close()
								}
								break
							}
							outBuf := new(bytes.Buffer)
							childCmd.SetOut(outBuf)
							if childCmd.Args != nil {
								if err := childCmd.Args(childCmd, flags); err != nil {
									err = client.Write([]byte(err.Error()))
									if err != nil {
										logger.Error("Write error", err)
										client.Close()
									}
									break
								}
							}
							if childCmd.PreRunE != nil {
								if err := childCmd.PreRunE(childCmd, flags); err != nil {
									err = client.Write([]byte(err.Error()))
									if err != nil {
										logger.Error("Write error", err)
										client.Close()
									}
									break
								}
							} else if childCmd.PreRun != nil {
								childCmd.PreRun(childCmd, command)
							}

							if childCmd.RunE != nil {
								if err := childCmd.RunE(childCmd, flags); err != nil {
									err = client.Write([]byte(err.Error()))
									if err != nil {
										logger.Error("Write error", err)
										client.Close()
									}
									break
								}
							} else if childCmd.Run != nil {
								childCmd.Run(childCmd, flags)
							}

							out := outBuf.Next(outBuf.Len())
							outBuf.Reset()
							err = client.Write(out)
							if err != nil {
								logger.Error("Write error", err)
								client.Close()
								break
							}
						}
					}()
				}
			}()
		}
	}
Execute:
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Flag definitions
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.stampkeeper.yaml)")
	rootCmd.PersistentFlags().StringVar(&verbosity, "verbosity", "5", "verbosity level")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".stampkeeper" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".stampkeeper")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If config file is not present, write it
	if err := viper.SafeWriteConfig(); err == nil {
		logger.Info("Using config file:", viper.ConfigFileUsed())
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Info("Using config file:", viper.ConfigFileUsed())
	}
}
