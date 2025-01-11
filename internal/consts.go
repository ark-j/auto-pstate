package internal

import (
	"fmt"
	"os"
)

const (
	defaultEppStateAC  = "balance_performance"
	defaultEppStateBat = "power"
	defaultGovernor    = "powersave"
	scalingDriverPath  = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_driver"
	governorPath       = "/sys/devices/system/cpu/cpu%d/cpufreq/scaling_governor"
	eppPath            = "/sys/devices/system/cpu/cpu%d/cpufreq/energy_performance_preference"
	powerPath          = "/sys/class/power_supply"
	allProfilesPath    = "/sys/devices/system/cpu/cpu0/cpufreq/energy_performance_available_preferences"
	allGovernorsPath   = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_available_governors"
)

const (
	AutoMode   = "auto"
	ManualMode = "manual"
)

var batPath = ""

func init() {
	batPath = GetPowerPath()
	if batPath == "" {
		fmt.Println("battery bath does not exist")
		os.Exit(1)
	}
}
