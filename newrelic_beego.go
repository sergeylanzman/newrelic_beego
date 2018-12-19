package newrelic_beego

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	newrelic "github.com/newrelic/go-agent"
)

var (
	reNumberIDInPath = regexp.MustCompile("[0-9]{2,}")
	reg              = regexp.MustCompile(`[a-zA-Z0-9_]+`)
	NewrelicAgent    newrelic.Application
)

const newRelicSkipPaths = "newrelic_skip_paths"
const newRelicAppName = "appname"
const newRelicLicense = "newrelic_license"

// Paths NewRelic needs to skip from reporting
var skipPaths map[string]bool

func init() {
	appName := os.Getenv("NEW_RELIC_APP_NAME")
	license := os.Getenv("NEW_RELIC_LICENSE_KEY")

	skipPaths = parseSkipPaths(beego.AppConfig.String(newRelicSkipPaths))

	if appName == "" {
		appName = beego.AppConfig.String("newrelic_appname")
	}
	if appName == "" {
		appName = beego.AppConfig.String(newRelicAppName)
	}
	if license == "" {
		license = beego.AppConfig.String(newRelicLicense)
		if license == "" && beego.BConfig.RunMode == "prod" {
			beego.Warn("Please set NewRelic license in config(newrelic_license)")
			return
		}
	}

	config := newrelic.NewConfig(appName, license)
	config.CrossApplicationTracer.Enabled = false
	app, err := newrelic.NewApplication(config)
	if err != nil {
		beego.Warn(err.Error())
		return
	}
	NewrelicAgent = app
	beego.InsertFilter("*", beego.BeforeRouter, StartTransaction, false)
	beego.InsertFilter("*", beego.AfterExec, NameTransaction, false)
	beego.InsertFilter("*", beego.FinishRouter, EndTransaction, false)
	beego.Info("NewRelic agent started")
}

// parseSkipPaths gets string of comma separated paths
// It returns a set of normalized paths
func parseSkipPaths(pathsConfig string) map[string]bool {
	paths := map[string]bool{}
	splitPaths := strings.Split(pathsConfig, ",")

	for _, path := range splitPaths {
		formatted := strings.TrimSpace(path)
		formatted = strings.ToLower(formatted)
		if formatted != "" {
			paths[formatted] = true
		}
	}

	return paths
}

const newRelicTransaction = "newrelic_transaction"

func StartTransaction(ctx *context.Context) {
	if shouldSkip(skipPaths, ctx.Request.URL.Path) {
		return
	}

	tx := NewrelicAgent.StartTransaction(ctx.Request.URL.Path, ctx.ResponseWriter.ResponseWriter, ctx.Request)
	ctx.ResponseWriter.ResponseWriter = tx
	ctx.Input.SetData(newRelicTransaction, tx)
}

// shouldSkip decides if given path matches against declared skip paths
// Support both exact matches and wildcard matches
func shouldSkip(skipPaths map[string]bool, path string) bool {
	_, exactMatch := skipPaths[path]

	if exactMatch {
		return true
	}

	for skipPath := range skipPaths {
		if strings.Contains(path, skipPath) {
			return true
		}
	}

	return false
}

func NameTransaction(ctx *context.Context) {
	var path string
	if ctx.Input.GetData(newRelicTransaction) == nil {
		return
	}
	tx := ctx.Input.GetData(newRelicTransaction).(newrelic.Transaction)
	// in old beego pattern available only in dev mode
	pattern, ok := ctx.Input.GetData("RouterPattern").(string)
	if ok {
		path = generatePath(pattern)
	} else {
		path = reNumberIDInPath.ReplaceAllString(ctx.Request.URL.Path, ":id")
	}

	displayExplicitEnv := beego.AppConfig.DefaultString("newrelic_display_explicit_env", "FALSE")
	if strings.ToUpper(displayExplicitEnv) == "TRUE" {
		env := strings.ToUpper(ctx.Input.Param(":env"))
		path = strings.Replace(path, ":env", env, -1)
	}

	txName := fmt.Sprintf("%s %s", ctx.Request.Method, path)
	_ = tx.SetName(txName)
}

func EndTransaction(ctx *context.Context) {
	if ctx.Input.GetData(newRelicTransaction) != nil {
		tx := ctx.Input.GetData(newRelicTransaction).(newrelic.Transaction)
		_ = tx.End()
	}
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
