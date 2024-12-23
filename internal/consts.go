package internal

const (
	defaultEppStateAC  = "balance_performance"
	defaultEppStateBat = "power"
	defaultGovernor    = "powersave"
	scalingDriverPath  = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_driver"
	governorPath       = "/sys/devices/system/cpu/cpu%d/cpufreq/scaling_governor"
	eppPath            = "/sys/devices/system/cpu/cpu%d/cpufreq/energy_performance_preference"
	batPath            = "/sys/class/power_supply/AC/online"
	allProfilesPath    = "/sys/devices/system/cpu/cpu0/cpufreq/energy_performance_available_preferences"
	allGovernorsPath   = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_available_governors"
)

const (
	AutoMode   = "auto"
	ManualMode = "manual"
)
