// Copyright 2017 Felipe A. Cavani. All rights reserved.
// Use of this source code is governed by the Apache License 2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	uetcd "github.com/fcavani/droute/etcd"
	drouterhttp "github.com/fcavani/droute/http"
	"github.com/fcavani/droute/middlewares/bucket"
	"github.com/fcavani/droute/responsewriter"
	"github.com/fcavani/droute/router"
	"github.com/fcavani/e"
	log "github.com/fcavani/slog"
	"github.com/fcavani/slog/systemd"
	"github.com/fcavani/systemd/watchdog"
	"github.com/fcavani/viperutil"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var daemonName = "drouter"

//Version stores the version number of this service.
var Version string

func init() {
	if Version == "" {
		Version = "dev"
	}
}

func main() {
	defer log.Recover(false)

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	viper.SetEnvPrefix(daemonName)
	viper.BindEnv("confdir")
	viper.SetConfigType("yaml")
	viper.SetConfigName("router")

	fset := flag.NewFlagSet("default", flag.ContinueOnError)
	help := fset.Bool("help", false, "Shows help.")
	ver := fset.Bool("version", false, "Show the version.")

	confdir := fset.String("confdir", ".", "Directory where will be the configuration files.")

	logLevel := fset.String("log-level", "", "Log level.")
	endpoints := fset.String("etcd-endpoints", "", "Etcd endpoints in a comma separated list.")
	secKeyring := fset.String("etcd-secring", "", "Etcd secret keyring to enable crypto store of the values.")
	etcdConfKey := fset.String("etcdkey", "/config/"+daemonName+".yaml", "Config file used to conf this software. This is the etcd key to retrieve de configuration.")
	name := fset.String("name", daemonName, "Name of the service")
	pidFile := fset.String("pid", daemonName+".pid", "Pid file for this service.")

	err := fset.Parse(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	if *help {
		fset.PrintDefaults()
		os.Exit(1)
	}
	if *ver {
		println(Version)
		os.Exit(1)
	}

	// Get the pid of this process to register it with etcd
	// pid := os.Getpid()
	// pidstr := strconv.FormatInt(int64(pid), 10)

	log.Tag("startup", "services", *name).Println("Starting:", *name)

	log.Tag("startup", "services", *name).Println("Watchdog...")

	ch, err := watchdog.Watchdog()
	if err != nil {
		log.Tag("startup", "services", *name).Error(err)
	} else {
		defer func() {
			ch <- struct{}{}
		}()
	}

	var eps []string
	var etc *uetcd.Etcd

	if *endpoints != "" {
		log.Tag("startup", "services", *name).Println("Configuring etcd...")
		eps = uetcd.LoadEtcdEndpoints(*endpoints)
		if *secKeyring == "" {
			viper.AddRemoteProvider("etcd", eps[0], *etcdConfKey)
		} else {
			viper.AddSecureRemoteProvider("etcd", eps[0], *etcdConfKey, *secKeyring)
		}
		err = viper.ReadRemoteConfig()
		if err != nil {
			log.Tag("startup", "services", *name).Fatalln(err)
		}

		etc = &uetcd.Etcd{
			Endpoints:  eps,
			SecKeyRing: *secKeyring,
		}
		err = etc.Init()
		if err != nil {
			log.Tag("startup", "services", *name).Fatal(err)
		}
	}

	if *confdir != "" {
		viper.AddConfigPath(*confdir)
		err = viper.ReadInConfig()
		if err != nil {
			log.Fatal("Can't read the configuration:", err)
		}
	}

	log.Tag("startup", "services", *name).Println("Configuring log level...")

	ll := viper.GetStringMapString("log")["level"]
	if *logLevel != "" {
		ll = *logLevel
	}
	level, err := log.ParseLevel(ll)
	if err != nil {
		log.Tag("startup", "services", *name).Fatalln(err)
	}
	setupLog(*name, level)

	if *pidFile != "" {
		log.Tag("startup", "services", *name).Println("Writing pid...")
		err = writePidFile(*pidFile)
		if err != nil {
			log.Tag("startup", "services", *name).Fatalln(err)
		}
	}

	// Create a type that will store all routes
	routers := router.NewRouters()
	// Get de default route, the only one so far.
	def := routers.Get(router.DefaultRouter)
	// Inject a signal in the http request.
	// The context will be caceled if one of the os signal came in, in this way,
	// the context will will have the opportunit to shutdown the http request.
	def.Context = func(ctx context.Context) (context.Context, context.CancelFunc) {
		return router.WithSignal(ctx, os.Interrupt, os.Kill)
	}
	// Add routes here.
	// Redir GET from anything to domain.com
	routers.Set("redir", router.NewRedirHostRouter("domain.com"))

	// LoadBalance strategy: round robin.
	lb := router.NewRoundRobin()

	// The router.
	r := &router.Router{}
	err = r.Start(routers, lb, 60*time.Second, 5)
	if err != nil {
		log.Tag("startup", "services", *name).Fatalln(err)
	}
	defer r.Stop()

	// HTTPHandlers example and bucket rate limit usage.
	r.HTTPHandlers(func(first http.Handler) http.Handler {
		return bucket.NewBucket(
			viperutil.GetMapN("http", "middleware", "bucket", "size").(int),
			time.Duration(viperutil.GetMapN("http", "middleware", "bucket", "timeout").(int))*time.Millisecond,
			first,
		)
	})

	// Middlewares example
	r.Middlewares(func(last responsewriter.HandlerFunc) responsewriter.HandlerFunc {
		return last
	})

	// Example of SetHostSwitch use...
	r.SetHostSwitch("domain.com", router.DefaultRouter)
	r.SetHostSwitch("www.domain.com", "redir")

	h := &drouterhttp.HTTPServer{
		HTTPAddr:           viper.GetStringMapString("http")["bindaddrs"],
		HTTPSAddr:          viper.GetStringMapString("https")["bindaddrs"],
		Certificate:        viper.GetStringMapString("https")["certificate"],
		PrivateKey:         viper.GetStringMapString("https")["privatekey"],
		CA:                 viper.GetStringMapString("https")["ca"],
		InsecureSkipVerify: viper.GetStringMap("https")["insecureskipverify"].(bool),
		Handler:            r,
	}

	err = h.Init()
	if err != nil {
		log.Tag("startup", "services", *name).Fatalln(err)
	}
	defer h.Stop()

	r.SetHTTPAddr(h.GetHTTPAddr())
	r.SetHTTPSAddr(h.GetHTTPSAddr())

	// a, err := host(lnHTTP.Addr())
	// if err != nil {
	// 	log.Tag("startup", "services", *name).Fatal(err)
	// }

	// Register
	// Open an unencrypted connection with etcd
	// var etcNE *uetcd.Etcd
	// if *endpoints != "" {
	// 	etcNE = &uetcd.Etcd{
	// 		Endpoints: eps,
	// 	}
	// 	err = etcNE.Init()
	// 	if err != nil {
	// 		log.Tag("startup", "services", *name).Fatal(err)
	// 	}
	// 	err = etcNE.Put(filepath.Join(common.Name, "http", pidstr, "addr"), []byte(a))
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	<-sig

	// if *endpoints != "" {
	// 	err = etcNE.Del(filepath.Join(daemonName, "pids", pidstr), &etcd.DeleteOptions{Recursive: true})
	// 	if err != nil {
	// 		log.Tag("startup", "services", *name).Fatal(err)
	// 	}
	// }

	if *pidFile != "" {
		err = os.Remove(*pidFile)
		if err != nil {
			log.Tag("startup", "services", *name).Fatal(err)
		}
	}
}

func setupLog(name string, level log.Level) {
	fname := viper.GetStringMapString("log")["file"]

	if fname == "" {
		log.Println("No log to file, log to stderr")
		log.SetOutput(name, level, os.Stderr, nil, nil, 1000)
		return
	}

	log.DebugLevel().Println("Open files...")

	f, err := os.OpenFile(fname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalln("open log file failed:", err)
	}

	log.DebugLevel().Println("Files opened...")
	log.DebugLevel().Println("Setup the logger...")

	if systemd.Enabled() {
		log.SetOutput(name, level, f, log.CommitSd, log.SdFormater, 1000)
	} else {
		log.SetOutput(name, level, f, nil, nil, 1000)
	}

	log.DebugInfo()

	log.Println("Logger configured...")
}

func writePidFile(file string) error {
	pid := os.Getpid()
	pidstr := strconv.FormatInt(int64(pid), 10)
	err := ioutil.WriteFile(file, []byte(pidstr), 0644)
	if err != nil {
		return e.Forward(err)
	}
	return nil
}

// func host(addr net.Addr) (string, error) {
// 	_, port, err := fnet.SplitHostPort(addr.String())
// 	if err != nil {
// 		return "", e.Forward(err)
// 	}
// 	return viperutil.GetMapNString("http", "bindAddrs") + ":" + port, nil
// }
