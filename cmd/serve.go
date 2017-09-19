package main

import (
	"context"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/color"
	"github.com/labstack/gommon/log"
	api "github.com/pintobikez/correios-service/api"
	uti "github.com/pintobikez/correios-service/config"
	strut "github.com/pintobikez/correios-service/config/structures"
	cronjob "github.com/pintobikez/correios-service/cronjob"
	lg "github.com/pintobikez/correios-service/log"
	rep "github.com/pintobikez/correios-service/repository/mysql"
	srv "github.com/pintobikez/correios-service/server"
	"github.com/robfig/cron"
	"gopkg.in/urfave/cli.v1"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var (
	repo       *rep.Repository
)

func init() {
	repo = new(rep.Repository)
}

// Start Http Server
func Serve(c *cli.Context) error {

	// Echo instance
	e := &srv.Server{echo.New()}
	e.HTTPErrorHandler = api.Error
	e.Logger.SetLevel(log.INFO)
	e.Logger.SetOutput(lg.File(c.String("log-folder") + "/app.log"))

	// Middlewares
	e.Use(middleware.LoggerWithOutput(lg.File(c.String("log-folder") + "/access.log")))
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
	err = repo.ConnectDB(stringConn)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer repo.DisconnectDB()

	//loads correios config
	correiosCnf, err := buildCorreiosConfig(c.String("correios-file"))
	if err != nil {
		e.Logger.Fatal(err)
	}

	a := api.New(repo, correiosCnf)
	cj := cronjob.New(repo, correiosCnf)

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

	if c.String("revision-file") != "" {
		e.File("/rev.txt", c.String("revision-file"))
	}

	if swagger := c.String("swagger-file"); swagger != "" {
		g := e.Group("/docs")
		g.Use(mw.CORSWithConfig(
			mw.CORSConfig{
				AllowOrigins: []string{"http://petstore.swagger.io"},
				AllowMethods: []string{echo.GET, echo.HEAD},
			},
		))

		g.GET("", func(c echo.Context) error {
			return c.File(swagger)
		})
	}

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

	// launch a cron to check everyday for posted items
	cr := cron.New()
	cr.AddFunc("* 0 */6 * * *", func() { cj.CheckUpdatedReverses("C") })     // checks for Colect updates
	cr.AddFunc("* 10 */6 * * *", func() { cj.CheckUpdatedReverses("A") })    // checks for Postage updates
	cr.AddFunc("* */20 * * * *", func() { cj.ReprocessRequestsWithError() }) // checks for Requests with Error and reprocesses them
	cr.Start()
	defer cr.Stop()

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
func buildCorreiosConfig(filename string) (*strut.CorreiosConfig, error) {
	t := new(strut.CorreiosConfig)
	err := uti.LoadCorreiosConfigFile(filename, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func buildStringConnection(filename string) (string, error) {
	t := new(strut.DbConfig)
	err := uti.LoadDBConfigFile(filename, t)
	if err != nil {
		return "", err
	}
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...&paramN=valueN]
	stringConn := t.Driver.User + ":" + t.Driver.Pw
	stringConn += "@tcp(" + t.Driver.Host + ":" + strconv.Itoa(t.Driver.Port) + ")"
	stringConn += "/" + t.Driver.Schema + "?charset=utf8"

	return stringConn, nil
}
