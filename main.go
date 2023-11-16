package main

import (
	"fmt"
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

var log *slog.Logger

func main() {
	IsRoot()
	IsPState()
	SetState()
}

func init() {
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	})
	log = slog.New(h)
}

// checks if script is run by root or not
func IsRoot() {
	if os.Geteuid() != 0 {
		log.Error("script must be run with root")
	}
}

// check if amd-pstate-driver present
func IsPState() {
	b, err := os.ReadFile(scalingDriverPath)
	if err != nil {
		log.Error("file does not exists for scaling driver", err)
	}
	if strings.TrimSpace(string(b)) != "amd-pstate-epp" {
		log.Error("system is not running amd-pstate-epp")
	}
}

// set the powersave governer if not already set
func SetGoverner() {
	b, err := os.ReadFile(fmt.Sprintf(governerPath, 0))
	if err != nil {
		log.Warn("governer file does not exists")
	}
	if string(b) != "powersave" {
		for i := 0; i < runtime.NumCPU(); i++ {
			if err := os.WriteFile(fmt.Sprintf(governerPath, i),
				[]byte("powersave"), os.ModePerm); err != nil {
				log.Error("while setting powersave governer", slog.Int("core", i), slog.String("err", err.Error()))
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
			log.Error("while setting epp_value", slog.String("value", val), slog.Int("core", i), slog.String("err", err.Error()))
			continue
		}
	}
}

// listens for he dbus event and sets the performance governer
func SetState() {
	// after first run whenever charhing changes so does our epp-state
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Error("unable to connect to system bus", err)
		return
	}
	defer conn.Close()

	signal := make(chan *dbus.Signal, 10)
	conn.Signal(signal)

	if err = conn.AddMatchSignal(
		dbus.WithMatchObjectPath(dbus.ObjectPath(upowerPath)),
	); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}

	for msg := range signal {
		if msg.Name == "org.freedesktop.DBus.Properties.PropertiesChanged" {
			m, ok := msg.Body[1].(map[string]dbus.Variant)
			if ok {
				switch m["Online"].Value().(bool) {
				case true:
					SetEPP(eppStateAC)
					log.Info("epp state set to balance_performance")
				case false:
					SetEPP(eppStateBat)
					log.Info("epp state set to power")
				}
			}

		}
	}
}
