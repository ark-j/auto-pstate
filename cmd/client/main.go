package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	var governor, profile, mode, timer string
	var listG, listP bool
	flag.StringVar(&mode, "m", "auto", modeDesc)
	flag.StringVar(&governor, "g", "powersave", governorDesc)
	flag.StringVar(&profile, "p", "balance_power", profileDesc)
	flag.StringVar(
		&timer,
		"t",
		"0.0",
		timerDesc,
	)
	flag.BoolVar(&listP, "list_profiles", false, listGDesc)
	flag.BoolVar(&listG, "list_governors", false, listPDesc)
	flag.Parse()

	if listG {
		for k := range internal.ListAvailable(true) {
			fmt.Println(k)
		}
	}

	if listP {
		for k := range internal.ListAvailable(false) {
			fmt.Println(k)
		}
	}
}

func ParseTime(s string) time.Duration {
	arr := strings.Split(s, ".")
	if len(arr) != 2 {
		fmt.Println("invalid format please enter in H.M or HH.MM format")
	}
	h, err := strconv.Atoi(arr[0])
	if err != nil {
		fmt.Println("invalid hour format")
		os.Exit(1)
	}
	m, err := strconv.Atoi(arr[1])
	if err != nil {
		fmt.Println("invalid minute format")
		os.Exit(1)
	}
	hour := time.Hour * time.Duration(h)
	minute := time.Minute * time.Duration(m)
	return hour + minute
}
