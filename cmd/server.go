/*
Copyright © 2020 NAME HERE christian@slashdevops.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/awshelper"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/collector"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/metrics"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/server"
	"github.com/slashdevops/aws_cloudwatch_exporter/internal/web"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// serverCmd represents the server command
var (
	serverCmd = &cobra.Command{
		Use:   "server [commands]",
		Short: "Useful to start server",
		Long:  `A longer description that spans `,
	}

	serverStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the http server to expose the metrics",
		Long:  `This start the http server to handle connections for metrics endpoint`,
		Run: func(cmd *cobra.Command, args []string) {
			startCmd(cmd, args)
		},
	}
)

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.AddCommand(serverStartCmd)

	// Server variables
	// Address
	serverCmd.PersistentFlags().StringVar(&conf.Server.Address, "address", "127.0.0.1", "Server address, empty means all addresses")
	if err := viper.BindPFlag("server.address", serverCmd.PersistentFlags().Lookup("address")); err != nil {
		log.Error(err)
	}

	// Port
	serverCmd.PersistentFlags().Uint16Var(&conf.Server.Port, "port", 9690, "Server port")
	if err := viper.BindPFlag("server.port", serverCmd.PersistentFlags().Lookup("port")); err != nil {
		log.Error(err)
	}

	// ReadTimeout
	serverCmd.PersistentFlags().DurationVar(&conf.Server.ReadTimeout, "readTimeout", 2*time.Second, "Server ReadTimeout in duration nomenclature, see https://golang.org/pkg/net/http/")
	if err := viper.BindPFlag("server.readTimeout", serverCmd.PersistentFlags().Lookup("readTimeout")); err != nil {
		log.Error(err)
	}

	// WriteTimeout
	serverCmd.PersistentFlags().DurationVar(&conf.Server.WriteTimeout, "writeTimeout", 5*time.Second, "Server WriteTimeout in duration nomenclature, see https://golang.org/pkg/net/http/")
	if err := viper.BindPFlag("server.writeTimeout", serverCmd.PersistentFlags().Lookup("writeTimeout")); err != nil {
		log.Error(err)
	}

	// IdleTimeout
	serverCmd.PersistentFlags().DurationVar(&conf.Server.IdleTimeout, "idleTimeout", 60*time.Second, "Server IdleTimeout in duration nomenclature, see https://golang.org/pkg/net/http/")
	if err := viper.BindPFlag("server.idleTimeout", serverCmd.PersistentFlags().Lookup("idleTimeout")); err != nil {
		log.Error(err)
	}

	// ReadHeaderTimeout
	serverCmd.PersistentFlags().DurationVar(&conf.Server.ReadHeaderTimeout, "readHeaderTimeout", 5*time.Second, "Server ReadHeaderTimeout in duration nomenclature, see https://golang.org/pkg/net/http/")
	if err := viper.BindPFlag("server.readHeaderTimeout", serverCmd.PersistentFlags().Lookup("readHeaderTimeout")); err != nil {
		log.Error(err)
	}

	// ShutdownTimeout
	serverCmd.PersistentFlags().DurationVar(&conf.Server.ShutdownTimeout, "shutdownTimeout", 30*time.Second, "Server ShutdownTimeout in duration nomenclature, the time waiting until graceful shutdown")
	if err := viper.BindPFlag("server.shutdownTimeout", serverCmd.PersistentFlags().Lookup("shutdownTimeout")); err != nil {
		log.Error(err)
	}

	// KeepAlivesEnabled
	serverCmd.PersistentFlags().BoolVar(&conf.Server.KeepAlivesEnabled, "keepAlivesEnabled", true, "Server KeepAlivesEnabled, see https://golang.org/pkg/net/http/")
	if err := viper.BindPFlag("server.keepAlivesEnabled", serverCmd.PersistentFlags().Lookup("keepAlivesEnabled")); err != nil {
		log.Error(err)
	}

	// LogFormat
	serverCmd.PersistentFlags().StringVar(&conf.Server.LogFormat, "logFormat", "text", "Server LogFormat, possible values: [text|json]")
	if err := viper.BindPFlag("server.logFormat", serverCmd.PersistentFlags().Lookup("logFormat")); err != nil {
		log.Error(err)
	}
}

func startCmd(cmd *cobra.Command, args []string) {
	ReadConfFromFiles()

	m := metrics.New(&conf)
	sess, _ := awshelper.NewSession(&conf.AWS)

	collector := collector.New(&conf, m, sess)
	prometheus.MustRegister(collector)

	handlers := web.NewHandlers(&conf)

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Home)
	mux.HandleFunc(conf.Application.HealthPath, handlers.Health)
	mux.Handle(conf.Application.MetricsPath, promhttp.Handler())

	// Debug & Profiling
	if conf.Server.Debug {
		mux.HandleFunc("/debug/pprof/", pprof.Index)
		mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	// this channel is to wait routines before finis this
	done := make(chan bool, 1)
	s := server.New(mux, &conf)

	// This run a go routine to listen Operating System signals
	// and execute a Gracefully shutdown when those occurs
	s.ListenOSSignals(&done)

	if err := s.Start(); err != nil {
		log.Fatalf("Server could not be started, %s", err.Error())
	}

	// Blocked until others routines finished
	<-done
}
