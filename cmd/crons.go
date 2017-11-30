package main

import (
	"fmt"
	"github.com/labstack/gommon/color"
	uti "github.com/pintobikez/brazilian-correios-service/config"
	strut "github.com/pintobikez/brazilian-correios-service/config/structures"
	cronjob "github.com/pintobikez/brazilian-correios-service/cronjob"
	lg "github.com/pintobikez/brazilian-correios-service/log"
	mysql "github.com/pintobikez/brazilian-correios-service/repository/mysql"
	"github.com/robfig/cron"
	"gopkg.in/urfave/cli.v1"
	"log"
)

//CronController Register a service in the Authentication Service and returns the generated API KEY
func CronController(c *cli.Context) error {

	f := lg.File(c.String("log-folder") + "/app.log")
	log.SetOutput(f)

	//loads db connection
	stringConn, err := buildStringConnection(c.String("database-file"))
	if err != nil {
		printErrorAndExit(err)
	}

	// Database connect
	repo := new(mysql.Client)
	err = repo.Connect(stringConn)
	if err != nil {
		printErrorAndExit(err)
	}
	defer repo.Disconnect()

	//loads correios config
	correiosCnf := new(strut.CorreiosConfig)
	err = uti.LoadConfigFile(c.String("correios-file"), correiosCnf)
	if err != nil {
		log.Fatal(err)
	}

	cj := cronjob.New(repo, correiosCnf)
	cj.SetOutput(f)

	// launch a cron to check everyday for posted items
	cr := cron.New()
	cr.AddFunc("* 0 */6 * * *", func() { cj.CheckUpdatedReverses("C") })     // checks for Colect updates
	cr.AddFunc("* 10 */6 * * *", func() { cj.CheckUpdatedReverses("A") })    // checks for Postage updates
	cr.AddFunc("* */20 * * * *", func() { cj.ReprocessRequestsWithError() }) // checks for Requests with Error and reprocesses them
	cr.Start()
	defer cr.Stop()

	fmt.Printf("%s %s\n", color.Green("[RESULT]"), "Cronjobs started.")

	return nil
}

func printErrorAndExit(err error) {
	log.Println(err.Error())
	fmt.Printf("%s %s\n", color.Red("[ERROR]"), err.Error())
	cli.OsExiter(1)
}
