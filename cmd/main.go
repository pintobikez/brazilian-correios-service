package main

import (
	"gopkg.in/urfave/cli.v1"
	"os"
)

var (
	appName = "brazilian-correios-service"
	version = "0.0.2"
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Version = version
	app.EnableBashCompletion = true
	app.Copyright = "(c) 2017 - Ricardo Pinto"
	app.Usage = "Correios service is a small app to deal Brazilian Correios return requests and track of objects"
	app.Commands = []cli.Command{
		//run as api
		cli.Command{
			Name:   "api",
			Usage:  "Runs the service as an api",
			Action: Handle,
		},
		// crons controller
		cli.Command{
			Name:   "cronjobs",
			Usage:  "Runs the crons needed for the service",
			Action: CronController,
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "log-folder, lf",
			Value:  "",
			Usage:  `Log folder path for access and application logging. Default "stdout"`,
			EnvVar: "LOG_FOLDER",
		},
		cli.StringFlag{
			Name:   "database-file, db",
			Value:  "",
			Usage:  "Database configuration `FILE` used by Correios Service to connect to database",
			EnvVar: "DATABASE_FILE",
		},
		cli.StringFlag{
			Name:   "correios-file, cr",
			Value:  "",
			Usage:  "Correios configuration `FILE` used by Correios Service to connect to Correios",
			EnvVar: "CORREIOS_FILE",
		},
		cli.StringFlag{
			Name:   "listen, l",
			Value:  "0.0.0.0:8000",
			Usage:  "Address and port on which Correios Service will accept HTTP requests (only for api)",
			EnvVar: "LISTEN",
		},
		cli.StringFlag{
			Name:   "ssl-cert, crt",
			Value:  "",
			Usage:  "Define SSL certificate to accept HTTPS requests (only for api)",
			EnvVar: "SSL_CERT",
		},
		cli.StringFlag{
			Name:   "ssl-key, key",
			Value:  "",
			Usage:  "Define SSL key to accept HTTPS requests (only for api)",
			EnvVar: "SSL_KEY",
		},
	}

	app.Run(os.Args)
}
