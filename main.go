package main

import (
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"github.com/z3ntl3/cf-uam-engine/api"
	filesystem "github.com/z3ntl3/cf-uam-engine/config"
	"github.com/z3ntl3/cf-uam-engine/worker"
)

func init() {
	log.SetTimeFormat(time.DateTime)
	log.SetReportCaller(true)

	style := lipgloss.NewStyle().
		Padding(0, 1, 0, 1).
		Foreground(lipgloss.Color("0"))

	styles := log.DefaultStyles()
	styles.Levels[log.InfoLevel] = style.
		Background(lipgloss.Color("117")).
		SetString("INFO")

	styles.Levels[log.ErrorLevel] = style.
		SetString("ERROR")

	log.SetStyles(styles)
	filesystem.ParseEnv()
}

func main() {
	var CONTROLLER = worker.NewController()

	c := api.New(viper.GetString("api_key"))
	if err := c.VerifyToken(); err != nil {
		log.Fatal(err)
	}
	log.Info("token is valid")

	domain, err := c.GetZone(viper.GetString("domain"))
	if err != nil {
		log.Fatal(err)
	}
	if len(domain.Result) == 0 {
		log.Fatalf("zone was not found for domain: %s", domain)
	}

	log.Infof("successfully found zone[%s]", domain.Result[0].ID)

	CONTROLLER.Client = c
	CONTROLLER.Loop(domain.Result[0].ID)
}
