package config

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/kore3lab/dashboard/backend/pkg/auth"
	"github.com/kore3lab/dashboard/backend/pkg/kubeconfig"
	"github.com/kore3lab/dashboard/backend/pkg/lang"
)

var Value = &conf{}
var Cluster *kubeCluster

var StartupOptions = struct {
	LogLevel    *string `json:"log-level"`
	TerminalUrl *string `json:"terminal-url"`
	KubeConfig  *string `json:"kubeconfig"`
	Auth        *string `json:"auth"`
}{}

var Authenticator *auth.Authenticator
var Clusters *kubeconfig.KubeContexts

func init() {

	// flags
	flag.StringVar(StartupOptions.LogLevel, "log-level", os.Getenv("LOG_LEVEL"), "The log level")
	flag.StringVar(StartupOptions.TerminalUrl, "terminal-url", os.Getenv("TERMINAL_URL"), "The address of the Terminal server")
	flag.StringVar(StartupOptions.KubeConfig, "kubeconfig", "", "The path to the kubeconfig used to connect to the Kubernetes API server and the Kubelets")
	flag.StringVar(StartupOptions.Auth, "auth", os.Getenv("AUTH"), "The authenticate options")

	//k8s.io client-go logs
	flag.Set("logtostderr", "ture")
	flag.Set("stderrthreshold", "FATAL")

	flag.Parse()

	//set default
	*StartupOptions.Auth = lang.NVL(*StartupOptions.Auth, "strategy=cookie,secret=static-token,token=kore3lab")
	*StartupOptions.TerminalUrl = lang.NVL(*StartupOptions.TerminalUrl, "http://localhost:3003")
	*StartupOptions.LogLevel = lang.NVL(*StartupOptions.LogLevel, "debug")

	//logger
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stderr)

	level, err := log.ParseLevel(*StartupOptions.LogLevel)
	if err != nil {
		log.Fatal(err)
	} else {
		log.SetLevel(level)
		log.Infof("Log level is '%s'", *StartupOptions.LogLevel)
	}

	// print startup options
	log.Infof("Startup options is '%v'", StartupOptions)

	// intialize "kubernetes-client"
	if Clusters, err := kubeconfig.NewKubeContexts(*StartupOptions.KubeConfig); err != nil {
		log.Errorf("Invalid a authenticator (cause=%s)", err.Error())
	} else {
		log.Infof("Initialized a kubernetes clusters (count=%s, current=%s)", len(Clusters.RESTConfigs), Clusters.CurrentContext)
	}

	// intialize "authenticator"
	if Authenticator, err = auth.NewAuthenticator(*StartupOptions.Auth); err != nil {
		log.Errorf("Invalid a authenticator (cause=%s)", err.Error())
	} else {
		log.Infof("Initialized a authenticator (count=%s, current=%s)")
	}

}
