package worker

import (
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
	"github.com/shirou/gopsutil/cpu"
	"github.com/spf13/viper"
	"github.com/z3ntl3/cf-uam-engine/api"
)

type Controller struct {
	UAM_Enabled *atomic.Bool
	Queue       chan int
	*api.Client
	*Settings
}

func NewController() *Controller {
	return &Controller{
		UAM_Enabled: &atomic.Bool{},
		Queue:       make(chan int, 1),
		Settings: &Settings{
			activate_treshold:  viper.GetFloat64("activate_treshold"),
			disengage_treshold: viper.GetFloat64("disengage_treshold"),
			uam_profile:        viper.GetString("uam_profile"),
			domain:             viper.GetString("domain"),
		},
	}
}

func (c *Controller) Loop(zoneId string) {
	CONTROLLER := c
	initial := false

	for {
		if initial {
			time.Sleep(time.Second * 10)
		} else {
			initial = true
		}

		go func() {
			load := cpu_load()
			isEnabled := CONTROLLER.UAM_Enabled.Load()
			defer func() {
				CONTROLLER.Queue <- 1
			}()

			if load >= CONTROLLER.activate_treshold && !isEnabled {
				if err := c.UpdateZone(CONTROLLER.uam_profile, zoneId); err != nil {
					log.Errorf("failed activating UAM: %s", err)
					return
				}

				CONTROLLER.UAM_Enabled.Store(true)
				log.Infof("under attack mode activated for %s", zoneId)
			} else if load <= CONTROLLER.disengage_treshold && isEnabled {
				if err := c.UpdateZone(CONTROLLER.uam_profile, zoneId); err != nil {
					log.Errorf("could not set UAM to low sensitivity: %s", err)
					return
				}

				CONTROLLER.UAM_Enabled.Store(false)
				log.Infof("under attack mode deactivated because load is below percentage for %s", zoneId)
			}

		}()

		<-CONTROLLER.Queue
	}
}

type Settings struct {
	activate_treshold  float64
	disengage_treshold float64
	domain             string
	uam_profile        string
}

func cpu_load() float64 {
	_, err := cpu.Percent(0, false)
	if err != nil {
		log.Fatal(err)
	}

	return 90.00
}
