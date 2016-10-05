package newrelic_beego

import (
	"fmt"
	"regexp"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/newrelic/go-agent"
)

var NewrelicAgent newrelic.Application

func StartTransaction(ctx *context.Context) {
	re := regexp.MustCompile("[0-9]{2,}")
	path := re.ReplaceAllString(ctx.Request.URL.Path, ":id")
	txName := fmt.Sprintf("%s %s", ctx.Request.Method, path)
	tx := NewrelicAgent.StartTransaction(txName, ctx.ResponseWriter, ctx.Request)
	ctx.Input.SetData("newrelic_transaction", tx)
}
func EndTransaction(ctx *context.Context) {
	if ctx.Input.GetData("newrelic_transaction") != nil {
		tx := ctx.Input.GetData("newrelic_transaction").(newrelic.Transaction)
		tx.End()
	}
}

func init() {
	appName := beego.AppConfig.String("newrelic_appname")
	if appName == "" {
		appName = beego.AppConfig.String("appname")
	}
	license := beego.AppConfig.String("newrelic_license")
	if license == "" {
		beego.Warn("Please set NewRelic license in config(newrelic_license)")
		return
	}
	config := newrelic.NewConfig(appName, license)
	app, err := newrelic.NewApplication(config)
	if err != nil {
		beego.Warn(err.Error())
		return
	}
	NewrelicAgent = app
	beego.InsertFilter("*", beego.BeforeRouter, StartTransaction, false)
	beego.InsertFilter("*", beego.FinishRouter, EndTransaction, false)
	beego.Info("NewRelic agent start")
}
