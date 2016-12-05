package newrelic_beego

import (
	"fmt"
	"regexp"
	"strings"

	"os"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/newrelic/go-agent"
)

var NewrelicAgent newrelic.Application

func StartTransaction(ctx *context.Context) {
	tx := NewrelicAgent.StartTransaction(ctx.Request.URL.Path, ctx.ResponseWriter, ctx.Request)
	ctx.Input.SetData("newrelic_transaction", tx)
}
func NameTransaction(ctx *context.Context) {
	var path string
	if ctx.Input.GetData("newrelic_transaction") == nil {
		return
	}
	tx := ctx.Input.GetData("newrelic_transaction").(newrelic.Transaction)
	// in old beego pattern available only in dev mode
	pattern, ok := ctx.Input.GetData("RouterPattern").(string)
	if ok {
		path = generatePath(pattern)
	} else {
		re := regexp.MustCompile("[0-9]{2,}")
		path = re.ReplaceAllString(ctx.Request.URL.Path, ":id")
	}
	txName := fmt.Sprintf("%s %s", ctx.Request.Method, path)
	tx.SetName(txName)
}
func EndTransaction(ctx *context.Context) {
	if ctx.Input.GetData("newrelic_transaction") != nil {
		tx := ctx.Input.GetData("newrelic_transaction").(newrelic.Transaction)
		tx.End()
	}
}

func init() {
	appName := os.Getenv("NEW_RELIC_APP_NAME")
	license := os.Getenv("NEW_RELIC_LICENSE_KEY")

	if appName == "" {
		appName = beego.AppConfig.String("newrelic_appname")
	}
	if appName == "" {
		appName = beego.AppConfig.String("appname")
	}
	if license == "" {
		license = beego.AppConfig.String("newrelic_license")
		if license == "" && beego.BConfig.RunMode == "prod" {
			beego.Warn("Please set NewRelic license in config(newrelic_license)")
			return
		}
	}
	config := newrelic.NewConfig(appName, license)
	app, err := newrelic.NewApplication(config)
	if err != nil {
		beego.Warn(err.Error())
		return
	}
	NewrelicAgent = app
	beego.InsertFilter("*", beego.BeforeRouter, StartTransaction, false)
	beego.InsertFilter("*", beego.BeforeExec, NameTransaction, false)
	beego.InsertFilter("*", beego.FinishRouter, EndTransaction, false)
	beego.Info("NewRelic agent start")
}

func generatePath(pattern string) string {
	segments := splitPath(pattern)
	for i, seg := range segments {
		segments[i] = replaceSegment(seg)
	}
	return strings.Join(segments, "/")
}
func splitPath(key string) []string {
	key = strings.Trim(key, "/ ")
	if key == "" {
		return []string{}
	}
	return strings.Split(key, "/")
}
func replaceSegment(seg string) string {
	colonSlice := []rune{':'}
	if strings.ContainsAny(seg, ":") {
		var newSegment []rune
		var start bool
		var startexp bool
		var param []rune
		var skipnum int
		reg := regexp.MustCompile(`[a-zA-Z0-9_]+`)
		for i, v := range seg {
			if skipnum > 0 {
				skipnum--
				continue
			}
			if start {
				//:id:int and :name:string
				if v == ':' {
					if len(seg) >= i+4 {
						if seg[i+1:i+4] == "int" {
							start = false
							startexp = false
							newSegment = append(newSegment, append(colonSlice, param...)...)
							skipnum = 3
							param = make([]rune, 0)
							continue
						}
					}
					if len(seg) >= i+7 {
						if seg[i+1:i+7] == "string" {
							start = false
							startexp = false
							newSegment = append(newSegment, append(colonSlice, param...)...)
							skipnum = 6
							param = make([]rune, 0)
							continue
						}
					}
				}
				// params only support a-zA-Z0-9
				if reg.MatchString(string(v)) {
					param = append(param, v)
					continue
				}
				if v != '(' {
					newSegment = append(newSegment, append(colonSlice, param...)...)
					param = make([]rune, 0)
					start = false
					startexp = false
				}
			}
			if startexp {
				if v != ')' {
					continue
				}
			}
			// Escape Sequence '\'
			if i > 0 && seg[i-1] == '\\' {
				newSegment = append(newSegment, v)
			} else if v == ':' {
				param = make([]rune, 0)
				start = true
			} else if v == '(' {
				startexp = true
				start = false
				if len(param) > 0 {
					newSegment = append(newSegment, append(colonSlice, param...)...)
					param = make([]rune, 0)
				}
			} else if v == ')' {
				startexp = false
				param = make([]rune, 0)
			} else if v == '?' {
				newSegment = append(newSegment, append([]rune{'?'}, param...)...)
			} else {
				newSegment = append(newSegment, v)
			}
		}
		if len(param) > 0 {
			newSegment = append(newSegment, append(colonSlice, param...)...)
		}
		return string(newSegment)
	}
	return seg
}
