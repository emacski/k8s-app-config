package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	metav1 "k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
)

var VERSION = "dev"

// error check
func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err)
		os.Exit(1)
	}
}

// return a ready-to-use kubernetes api client
func kubeClient() *kubernetes.Clientset {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	check(err)
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	check(err)
	return clientset
}

// hosts returns a slice of strings where each element is an ip address
// of a host registered to the supplied kubernetes namespace / service.
func hosts(client *kubernetes.Clientset, namespace, service string, wait, nodes int) ([]string, error) {
	addrs := []string{}
	var err error
	// flattenSubsets function from:
	// https://github.com/kubernetes/kubernetes/blob/master/cluster/addons/fluentd-elasticsearch/es-image/elasticsearch_logging_discovery.go
	// Copyright 2017 The Kubernetes Authors. http://www.apache.org/licenses/LICENSE-2.0
	var flattenSubsets = func(subsets []metav1.EndpointSubset) []string {
		ips := []string{}
		for _, ss := range subsets {
			for _, addr := range ss.Addresses {
				ips = append(ips, addr.IP)
			}
		}
		return ips
	}
	if wait > 0 {
		for t := time.Now(); time.Since(t) < time.Duration(wait)*time.Minute; time.Sleep(10 * time.Second) {
			endpoints, err := client.CoreV1().Endpoints(namespace).Get(service)
			if err != nil {
				continue
			}
			addrs = flattenSubsets(endpoints.Subsets)
			if len(addrs) >= nodes {
				break
			}
		}
		if err != nil {
			return nil, err
		}
	} else {
		endpoints, err := client.CoreV1().Endpoints(namespace).Get(service)
		if err != nil {
			return nil, err
		}
		addrs = flattenSubsets(endpoints.Subsets)
	}
	return addrs, nil
}

// format output using a template file, output is to stdout
func fmtTemplateFile(tplPath string, tplVars map[string]interface{}) error {
	tpl, err := ioutil.ReadFile(tplPath)
	if err != nil {
		return err
	}
	return fmtTemplateString(string(tpl), tplVars)
}

// format output using a template string, output is to stdout
func fmtTemplateString(tpl string, tplVars map[string]interface{}) error {
	t, err := template.New("").Funcs(template.FuncMap{
		"join": strings.Join, // additional template functions
	}).Parse(tpl)
	if err != nil {
		return err
	}
	err = t.Execute(os.Stdout, tplVars)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var hostsNamespc, hostsService, hostsFormat, hostsTemplate string
	var hostsWait, hostsMinHosts int
	var cmdHosts = &cobra.Command{
		Use:   "hosts",
		Short: "Output host ips for a give namespace and service",
		Run: func(cmd *cobra.Command, args []string) {
			h, err := hosts(kubeClient(), hostsNamespc, hostsService, hostsWait, hostsMinHosts)
			check(err)
			if hostsTemplate == "" && hostsFormat == "" {
				fmt.Println(h)
			} else if hostsFormat != "" {
				check(fmtTemplateString(hostsFormat, map[string]interface{}{"hosts": h}))
			} else {
				check(fmtTemplateFile(hostsTemplate, map[string]interface{}{"hosts": h}))
			}
		},
	}
	cmdHosts.Flags().StringVarP(&hostsNamespc, "namespace", "n", "", "k8s namespace of application")
	cmdHosts.Flags().StringVarP(&hostsService, "service", "s", "", "k8s service name of application")
	cmdHosts.Flags().StringVarP(&hostsFormat, "format", "f", "", "go template string to use for output formatting")
	cmdHosts.Flags().StringVarP(&hostsTemplate, "template", "t", "", "path to go template file to use for output formatting")
	cmdHosts.Flags().IntVarP(&hostsWait, "wait", "w", 0, "how long to wait (in minutes) for --min-nodes")
	cmdHosts.Flags().IntVarP(&hostsMinHosts, "min-hosts", "m", 2, "minimum number of hosts to --wait for")

	var ver bool
	var rootCmd = &cobra.Command{
		Use:   os.Args[0],
		Short: "Kubernetes application config helper utility",
		Run: func(cmd *cobra.Command, args []string) {
			if ver {
				fmt.Printf("%s version %s\n", os.Args[0], VERSION)
			}
		},
	}
	rootCmd.Flags().BoolVar(&ver, "version", false, "print version")
	rootCmd.AddCommand(cmdHosts)
	rootCmd.Execute()
}
