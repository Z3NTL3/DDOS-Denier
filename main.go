package main

import (
	"log"
	"os"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/spf13/viper"
	"github.com/z3ntl3/cf-uam-engine/api"
	"github.com/z3ntl3/cf-uam-engine/filesystem"
)

func cpu_load() float64 {
	avg, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Fatal(err)
	}

	return avg[0]
}

var UAM_enabled = false

func main() {
	filesystem.ParseEnv()

	logger := log.New(os.Stdout, "[LOG]: ", log.Ltime)
	c := api.New(viper.GetString("apiKey"))
	if err := c.VerifyToken(); err != nil {
		log.Fatal(err)
	}
	logger.Printf("token is valid")

	domain, err := c.GetZone(viper.GetString("domain"))
	if err != nil {
		log.Fatal(err)
	}
	logger.Printf("successfully found zone[%s]", domain.Result[0].ID)

	init := false
	for {
		if init {
			time.Sleep(time.Second * 2)
		} else {
			init = true
		}

		lookupCycle := make(chan int, 1)

		go func(cyc chan int, isEnabled *bool) {
			load := cpu_load()

			if load >= viper.GetFloat64("activateAfter") && !*isEnabled {
				if err := c.UpdateZone(viper.GetString("UAM"), domain.Result[0].ID); err != nil {
					logger.Printf("failed activating UAM: %s", err)
				}

				*isEnabled = true
				logger.Printf("under attack mode activated for %s", domain.Result[0].ID)
			} else if load <= viper.GetFloat64("closeBelow") && *isEnabled {
				if err := c.UpdateZone(viper.GetString("LOW"), domain.Result[0].ID); err != nil {
					logger.Printf("could not set UAM to low sensitivity: %s", err)
				}

				*isEnabled = false
				logger.Printf("under attack mode deactivated because load is below percentage for %s", domain.Result[0].ID)
			}

			cyc <- 1
		}(lookupCycle, &UAM_enabled)

		<-lookupCycle
	}
}
