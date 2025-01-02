package internal

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// polling sysfs it checks wether devices on charging
func charging() bool {
	b, err := os.ReadFile(batPath)
	if err != nil {
		slog.Warn("file not found for AC")
		os.Exit(1)
	}
	return strings.TrimSpace(string(b)) == "1"
}

// checks if script is run by root or not
func IsRoot() {
	if os.Geteuid() != 0 {
		slog.Error("script must be run with root")
		os.Exit(1)
	}
}

// list all governors or profiles
func ListAvailable(profile bool) map[string]bool {
	var (
		b   []byte
		err error
	)
	if profile {
		b, err = os.ReadFile(allProfilesPath)
	} else {
		b, err = os.ReadFile(allGovernorsPath)
	}
	if err != nil {
		slog.Error(err.Error())
		return nil
	}
	m := make(map[string]bool)
	for _, f := range strings.Fields(strings.TrimSpace(string(b))) {
		m[f] = true
	}
	return m
}

// check if amd-pstate-driver present
func IsPState() {
	b, err := os.ReadFile(scalingDriverPath)
	if err != nil {
		slog.Error("file does not exists for scaling driver", slog.String("err", err.Error()))
		os.Exit(1)
	}
	if strings.TrimSpace(string(b)) != "amd-pstate-epp" {
		slog.Error("system is not running amd-pstate-epp")
		os.Exit(1)
	}
}

func ParseTime(s string) (d time.Duration) {
	arr := strings.Split(s, ":")
	if len(arr) == 2 {
		h, err := strconv.Atoi(arr[0])
		if err != nil {
			slog.Error("invalid format", slog.String("err", err.Error()))
			os.Exit(1)
		}
		if h > 24 || h < 0 {
			slog.Error("invalid format", slog.String("err", "hour format is wrong"))
			os.Exit(1)
		}
		m, err := strconv.Atoi(arr[1])
		if err != nil {
			slog.Error("invalid format", slog.String("err", err.Error()))
			os.Exit(1)
		}
		if m > 60 || m < 0 {
			slog.Error("invalid format", slog.String("err", "minute format is wrong"))
			os.Exit(1)
		}
		d = time.Duration(h)*time.Hour + time.Duration(m)*time.Minute
	} else {
		slog.Error("invalid format")
	}
	return
}

// H is used to send dynamic response
type H map[string]any

// JSON response utils
func JSON(b any, w http.ResponseWriter, status int) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(b)
}

// Bind json body to struct
func Bind(r io.Reader, m any) error {
	return json.NewDecoder(r).Decode(m)
}
