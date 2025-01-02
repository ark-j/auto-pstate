package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ark-j/auto-pstate/internal"
)

const (
	modeDesc     = "mode can be auto/manual"
	governorDesc = "set governor which can be powersave, performance"
	profileDesc  = "set profile you can list them using list-profiles"
	timerDesc    = `timer in format H.M or HH.MM
timer will reset mode to auto after elapse of set time.
ex. 1.5 will set timer for 1 hour 15 minutes 1.0 will set for 1 hour and 0.15 will set 15 minutes`
	listGDesc = "list all profiles available"
	listPDesc = "list all governor available"
)

func main() {
	var governor, profile, mode string
	var listG, listP bool
	flag.StringVar(&mode, "m", "auto", modeDesc)
	flag.StringVar(&governor, "g", "powersave", governorDesc)
	flag.StringVar(&profile, "p", "balance_power", profileDesc)
	flag.BoolVar(&listP, "list_profiles", false, listGDesc)
	flag.BoolVar(&listG, "list_governors", false, listPDesc)
	flag.Parse()

	if listG {
		for k := range internal.ListAvailable(true) {
			fmt.Println(k)
		}
		return
	}

	if listP {
		for k := range internal.ListAvailable(false) {
			fmt.Println(k)
		}
		return
	}

	if mode == internal.AutoMode {
		Auto()
		return
	}

	if !internal.ListAvailable(false)[governor] {
		fmt.Println("invalid governor")
		os.Exit(1)
	}

	if !internal.ListAvailable(true)[profile] {
		fmt.Println("invalid profile")
		os.Exit(1)
	}

	if mode == internal.ManualMode {
		Manual(governor, profile)
	}
}

var client *http.Client

func init() {
	client = &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _ string, _ string) (net.Conn, error) {
				return net.Dial("unix", internal.SockPath)
			},
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 30 * time.Second,
	}
}

func Manual(g, p string) {
	b, _ := json.Marshal(internal.ManualRequest{ //nolint
		Governor: g,
		Profile:  p,
	})
	res, err := client.Post("http://localhost:3003/epp/manual", "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer res.Body.Close()

	b, err = io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if res.Header.Get("Content-Type") == "application/json" && res.StatusCode == http.StatusOK {
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(m["msg"])
		return
	}
	fmt.Println(string(b))
}

func Auto() {
	res, err := client.Post("http://localhost:3003/epp/auto", "application/json", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if res.Header.Get("Content-Type") == "application/json" && res.StatusCode == http.StatusOK {
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(m["msg"])
		return
	}
	fmt.Println(string(b))
}
