newrelic_beego
=============

newrelic_beego is "plug and play" package for monitoring(APM) beego framework with newrelic offical agent

Support http endpoints
 
Can Get newrelic_beego.NewrelicAgent for custom monitoring(database,external call, func etc..)
Can Get newrelic transaction per request from beego context
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
go get github.com/sergeylanzman/newrelic_beego"
```

Add  _ "github.com/sergeylanzman/newrelic_beego" to import in main.go file

# Settings
    - appname = name of app in newrelic
    - newrelic_license = newrelic license
