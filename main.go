package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	eppStateAC        = "balance_performance"
	eppStateBat       = "power"
	scalingDriverPath = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_driver"
	governerPath      = "/sys/devices/system/cpu/cpu%d/cpufreq/scaling_governor"
	eppPath           = "/sys/devices/system/cpu/cpu%d/cpufreq/energy_performance_preference"
	batPath           = "/sys/class/power_supply/AC/online"
)

func main() {
	IsRoot()
	IsPState()
	// sets the state variable to opposite of charging for first run
	state := false
	if !Charging() {
		state = true
	}
	// after first run whenever charhing changes so does our epp-state
	for true {
		if state != Charging() {
			SetState()
			state = Charging()
		}
		time.Sleep(5 * time.Second)
	}
}

// checks if script is run by root or not
func IsRoot() {
	if os.Geteuid() != 0 {
		log.Fatal("[ERROR] script must be run with root")
	}
}

// check if amd-pstate-driver present
func IsPState() {
	b, err := os.ReadFile(scalingDriverPath)
	if err != nil {
		log.Fatal("[WARNING] file does not exists for scaling driver", err)
	}
	if strings.TrimSpace(string(b)) != "amd-pstate-epp" {
		log.Fatal("[WARNING] system is not running amd-pstate-epp")
	}
}

// check if system on charging or on battery
func Charging() bool {
	b, err := os.ReadFile(batPath)
	if err != nil {
		log.Fatal("[WARNING] file not found for AC")
	}
	return strings.TrimSpace(string(b)) == "1"
}

// set the pwoersave governer if not already set
func SetGoverner() {
	b, err := os.ReadFile(fmt.Sprintf(governerPath, 0))
	if err != nil {
		log.Println("[WARNING] governer file does not exists")
	}
	if string(b) != "powersave" {
		for i := 0; i < runtime.NumCPU(); i++ {
			if err := os.WriteFile(fmt.Sprintf(governerPath, i),
				[]byte("powersave"), os.ModePerm); err != nil {
				log.Printf("[ERROR] while setting powersave governer to cpu core %d err -> %v\n", i, err)
				continue
			}
		}
	}
}

// set epp value for performance and power consumption
func SetEPP(val string) {
	for i := 0; i < runtime.NumCPU(); i++ {
		err := os.WriteFile(fmt.Sprintf(eppPath, i), []byte(val), os.ModePerm)
		if err != nil {
			log.Printf("[ERROR] while setting %s epp_value to cpu core %d err -> %v\n", val, i, err)
			continue
		}
	}
}

// set the proper state based on charging
func SetState() {
	if Charging() {
		SetEPP(eppStateAC)
		fmt.Println("[INFO] epp state set to balance_performance")
	} else {
		SetEPP(eppStateBat)
		fmt.Println("[INFO] epp state set to power")
	}
}
