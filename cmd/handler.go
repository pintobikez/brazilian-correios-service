package main

import (
	"context"
	"fmt"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/color"
	"github.com/labstack/gommon/log"
	api "github.com/pintobikez/brazilian-correios-service/api"
	uti "github.com/pintobikez/brazilian-correios-service/config"
	strut "github.com/pintobikez/brazilian-correios-service/config/structures"
	lg "github.com/pintobikez/brazilian-correios-service/log"
	mysql "github.com/pintobikez/brazilian-correios-service/repository/mysql"
	srv "github.com/pintobikez/brazilian-correios-service/server"
	"gopkg.in/urfave/cli.v1"
	"os"
	"os/signal"
	"strconv"
	"time"
)

//Handle Start Http Server
func Handle(c *cli.Context) error {

	fmt.Printf("Flags %d %v", c.NumFlags(), c.GlobalFlagNames())

	// Echo instance
	e := &srv.Server{echo.New()}
	e.HTTPErrorHandler = api.Error
	e.Logger.SetLevel(log.INFO)
	e.Logger.SetOutput(lg.File(c.String("log-folder") + "/app.log"))

	// Middlewares
	e.Use(lg.LoggerWithOutput(lg.File(c.String("log-folder") + "/access.log")))
	e.Use(mw.Recover())
	e.Use(mw.Secure())
	e.Use(mw.RequestID())
	e.Pre(mw.RemoveTrailingSlash())

	//loads db connection
	stringConn, err := buildStringConnection(c.String("database-file"))
	if err != nil {
		e.Logger.Fatal(err)
	}

	// Database connect
	repo := new(mysql.Client)
	err = repo.Connect(stringConn)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer repo.Disconnect()

	//loads correios config
	correiosCnf := new(strut.CorreiosConfig)
	err = uti.LoadConfigFile(c.String("correios-file"), correiosCnf)
	if err != nil {
		e.Logger.Fatal(err)
	}

	a := api.New(repo, correiosCnf)

	// Routes => api
	e.POST("/tracking", a.GetTracking())
	e.POST("/reverse", a.PostReverse())
	e.POST("/reversesearch", a.GetReversesBy())
	e.PUT("/reverse/:requestId", a.PutReverse())
	e.DELETE("/reverse/:requestId", a.DeleteReverse())
	e.GET("/reverse/:requestId", a.GetReverse())
	e.Use(mw.CORSWithConfig(
		mw.CORSConfig{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{echo.POST, echo.GET, echo.DELETE, echo.PUT, echo.HEAD},
		},
	))

	// Start server
	colorer := color.New()
	colorer.Printf("⇛ %s service - %s\n", appName, color.Green(version))
	//Print available routes
	colorer.Printf("⇛ Available Routes:\n")
	for _, rou := range e.Routes() {
		colorer.Printf("⇛ URI: [%s] %s\n", color.Green(rou.Method), color.Green(rou.Path))
	}

	go func() {
		if err := start(e, c); err != nil {
			colorer.Printf(color.Red("⇛ shutting down the server\n"))
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	return nil
}

// Start http server
func start(e *srv.Server, c *cli.Context) error {

	if c.String("ssl-cert") != "" && c.String("ssl-key") != "" {
		return e.StartTLS(
			c.String("listen"),
			c.String("ssl-cert"),
			c.String("ssl-key"),
		)
	}

	return e.Start(c.String("listen"))
}

func buildStringConnection(filename string) (string, error) {
	t := new(strut.DbConfig)
	if err := uti.LoadConfigFile(filename, t); err != nil {
		return "", err
	}
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	stringConn := t.Driver.User + ":" + t.Driver.Pw
	stringConn += "@tcp(" + t.Driver.Host + ":" + strconv.Itoa(t.Driver.Port) + ")"
	stringConn += "/" + t.Driver.Schema + "?charset=utf8"

	return stringConn, nil
}
