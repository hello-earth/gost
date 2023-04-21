package main

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"strings"

	_ "net/http/pprof"

	"github.com/ginuerzh/gost"
	"github.com/go-log/log"
)

var (
	configureFile string
	baseCfg       = &baseConfig{}
	pprofAddr     string
	pprofEnabled  = os.Getenv("PROFILING") != ""
)

func init() {
	gost.SetLogger(&gost.LogLogger{})

	var (
		printVersion bool
	)

	flag.Var(&baseCfg.route.ChainNodes, "F", "forward address, can make a forward chain")
	flag.Var(&baseCfg.route.ServeNodes, "L", "listen address, can listen on multiple ports (required)")
	flag.IntVar(&baseCfg.route.Mark, "M", 0, "Specify out connection mark")
	flag.StringVar(&configureFile, "C", "", "configure file")
	flag.StringVar(&baseCfg.route.Interface, "I", "", "Interface to bind")
	flag.BoolVar(&baseCfg.Debug, "D", false, "enable debug log")
	flag.StringVar(&baseCfg.Ccf, "CCF", "", "if replace a cloudflare domain to a better edge ip")
	flag.BoolVar(&printVersion, "V", false, "print version")
	if pprofEnabled {
		flag.StringVar(&pprofAddr, "P", ":6060", "profiling HTTP server address")
	}
	flag.Parse()

	if printVersion {
		fmt.Fprintf(os.Stdout, "gost %s (%s %s/%s)\n",
			gost.Version, runtime.Version(), runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	if configureFile != "" {
		_, err := parseBaseConfig(configureFile)
		if err != nil {
			log.Log(err)
			os.Exit(1)
		}
	}
	if flag.NFlag() == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}
	if len(baseCfg.Ccf) > 10 && strings.Contains(baseCfg.Ccf, "@") {
		filePath := strings.Split(baseCfg.Ccf, "@")[1]
		baseCfg.Ccf = strings.Split(baseCfg.Ccf, "@")[0]
		f, err := os.Open(filePath)
		if err != nil {
			fmt.Println("read file fail", err)
			os.Exit(0)
		}
		defer f.Close()

		fd, err := ioutil.ReadAll(f)
		if err != nil {
			fmt.Println("read to fd fail", err)
			os.Exit(0)
		}

		cfIps := string(fd)
		defaultCloudflareIPs = strings.Split(cfIps, "\n")
	}

}

func main() {
	if pprofEnabled {
		go func() {
			log.Log("profiling server on", pprofAddr)
			log.Log(http.ListenAndServe(pprofAddr, nil))
		}()
	}

	// NOTE: as of 2.6, you can use custom cert/key files to initialize the default certificate.
	tlsConfig, err := tlsConfig(defaultCertFile, defaultKeyFile, "")
	if err != nil {
		// generate random self-signed certificate.
		cert, err := gost.GenCertificate()
		if err != nil {
			log.Log(err)
			os.Exit(1)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	} else {
		log.Log("load TLS certificate files OK")
	}

	gost.DefaultTLSConfig = tlsConfig

	if err := start(); err != nil {
		log.Log(err)
		os.Exit(1)
	}

	select {}
}

func start() error {
	gost.Debug = baseCfg.Debug
	gost.CloudflareIPs = defaultCloudflareIPs
	gost.Cfdomain = baseCfg.Ccf

	var routers []router
	rts, err := baseCfg.route.GenRouters()
	if err != nil {
		return err
	}
	routers = append(routers, rts...)

	for _, route := range baseCfg.Routes {
		rts, err := route.GenRouters()
		if err != nil {
			return err
		}
		routers = append(routers, rts...)
	}

	if len(routers) == 0 {
		return errors.New("invalid config")
	}
	for i := range routers {
		go routers[i].Serve()
	}

	return nil
}
