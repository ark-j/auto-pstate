package main

import "flag"

func main() {
	var governor, profile, mode, timer string
	var listG, listP bool
	flag.StringVar(&mode, "m", "auto", "mode can be auto/manual")
	flag.StringVar(&governor, "g", "powersave", "set governor which can be powersave, performance")
	flag.StringVar(&profile, "p", "based on charging", "set profile you can list them using list-profiles")
	flag.StringVar(
		&timer,
		"t",
		"00:00",
		"timer in format HH:MM. timer will reset mode to auto after elapse of set time",
	)
	flag.BoolVar(&listP, "list_profiles", false, "list all profiles available")
	flag.BoolVar(&listG, "list_governors", false, "list all governor available")
	flag.Parse()
}
