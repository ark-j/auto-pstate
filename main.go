package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/godbus/dbus/v5"
)

const (
	eppStateAC        = "balance_performance"
	eppStateBat       = "power"
	scalingDriverPath = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_driver"
	governerPath      = "/sys/devices/system/cpu/cpu%d/cpufreq/scaling_governor"
	eppPath           = "/sys/devices/system/cpu/cpu%d/cpufreq/energy_performance_preference"
	upowerPath        = "/org/freedesktop/UPower/devices/line_power_AC"
)

func main() {
	IsRoot()
	IsPState()
	SetState()
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
		slog.Error("file does not exists for scaling driver", err)
	}
	if strings.TrimSpace(string(b)) != "amd-pstate-epp" {
		slog.Error("system is not running amd-pstate-epp")
	}
}

// set the powersave governer if not already set
func SetGoverner() {
	b, err := os.ReadFile(fmt.Sprintf(governerPath, 0))
	if err != nil {
		slog.Warn("governer file does not exists")
	}
	if string(b) != "powersave" {
		for i := 0; i < runtime.NumCPU(); i++ {
			if err := os.WriteFile(fmt.Sprintf(governerPath, i),
				[]byte("powersave"), os.ModePerm); err != nil {
				slog.Error(fmt.Sprintf("while setting powersave governer to cpu core %d err -> %v\n", i, err))
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

// listens for he dbus event and sets the performance governer
func SetState() {
	// after first run whenever charhing changes so does our epp-state
	conn, err := dbus.SystemBus()
	if err != nil {
		slog.Error("unable to connect to system bus", err)
		return
	}
	defer conn.Close()

	signal := make(chan *dbus.Signal, 10)
	conn.Signal(signal)

	if err = conn.AddMatchSignal(
		dbus.WithMatchObjectPath(dbus.ObjectPath(upowerPath)),
	); err != nil {
		log.Fatal(err)
	}

	for msg := range signal {
		if msg.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			m, ok := msg.Body[1].(map[string]dbus.Variant)
			if ok {
				switch m["Online"].Value().(bool) {
				case true:
					SetEPP(eppStateAC)
					fmt.Println("[INFO] epp state set to balance_performance")
				case false:
					SetEPP(eppStateBat)
					fmt.Println("[INFO] epp state set to power")
				}
			}

		}
	}
}
