[![Build Status](https://travis-ci.com/gtforge/newrelic_beego.svg?branch=master)](https://travis-ci.com/gtforge/newrelic_beego)

NewRelic BeeGo
==============

NewRelic BeeGo is "plug and play" package for monitoring (APM) BeeGo framework with NewRelic official agent<br />

Supports HTTP endpoints
 
You can use exposed newrelic_beego.NewrelicAgent for custom monitoring, such as database, external calls, functions, etc.

Also, you can get NewRelic transaction per request from BeeGo context:
```
txn := ctx.Input.GetData("newrelic_transaction").(newrelic.Transaction)
defer txn.EndDatastore(txn.StartSegment(), datastore.Segment{
    // Product is the datastore type.
    // See the constants in api/datastore/datastore.go.
    Product: datastore.MySQL,
    // Collection is the table or group.
    Collection: "my_table",
    // Operation is the relevant action, e.g. "SELECT" or "GET".
    Operation: "SELECT",
})
```

# Installation
```
dep ensure -v
```

Add  `_ "github.com/gtforge/newrelic_beego"` as an import in `main.go` file

# Available settings
- appname = name of app in newrelic
- newrelic_appname = same as `appname`
- newrelic_license = NewRelic license key
- newrelic_display_explicit_env = TRUE will display `RU/IL/UK` in the URL. FALSE will display just `:env`
- newrelic_skip_paths = comma separated paths that shouldn't be logged by NewRelic.
    - Example: `/api/v1/dosomething, /debug/`
    - Note: matching is fuzzy, with above example any path that contains `/debug/` will be skipped
