package discord

import (
	"bytes"

	"github.com/urfave/cli"
)

var buf = &bytes.Buffer{}
var app *cli.App

func init() {

	app = cli.NewApp()
	app.Name = "KRinder Discord Commands"
	app.Writer = buf
}
