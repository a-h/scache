# scache

A cache designed for Serverless applications which is invalidated by a Kinesis stream and keeps track of how effective it is.

## Details

One of the problems with cache is that after a database update occurs, the next HTTP request that retrieves that data can receive the data from the cache, not the latest version.

On a single server, it's easy enough to remove the key from the local cache when the data changes.

For load balanced servers, a common solution is to use a server such as Redis, and to remove the key from Redis when the data is updated. This way, all Web servers in a load balanced set will see updated data.

Unfortunately, AWS Serverless applications can't really do this, since Elasticache (AWS's managed Redis offering) needs to run in a VPC, and running in a VPC incurs a 10 second startup penalty on the Lambda while the Elastic Network Interface is attached. Introducing delay renders having a cache useless.

After a Lambda is executed, its state is persisted for up to a few hours allowing data to be cached in RAM, but how can a Lambda know when it starts back up, whether the data it has in RAM is now out-of-date? Any other Lambda running could have invalidated the data.

Kinesis streams offer a potential solution. When data is updated, the cache can be invalidated by writing a message to the Kinesis stream, each HTTP request just needs to check whether it should remove any items from the cache before using the cache (if there's anything in the cache).

Getting data from Kinesis takes time, so `scache` keeps track of how much time was spent on reading invalidation messages from Kinesis to compare to the value offered by the cache and writes it to a log.

# How to use it

## Full example

* Step 1: Setting up Middleware
    * [example/api/main.go](example/api/main.go)
* Step 2: Using it
    * [example/user/handler.go](example/user/handler.go)

## Add the HTTP middleware

```go
streamName := os.Getenv("KINESIS_STREAM_NAME")
if streamName == "" {
    log.Fatal("unable to start up, environment variable KINESIS_STREAM_NAME is missing")
}
stream := expiry.NewStream(streamName)
// Cache for between 30 seconds and 1 minute.
minCacheDuration, maxCacheDuration := time.Second*30, time.Minute
h := scache.AddMiddleware(next, stream, minCacheDuration, maxCacheDuration)
// Listen on port 8080.
http.ListenAndServe(":8080", h)
```

## Add items to the cache

```go
user := db.GetUser("12345")
dataID := data.NewID("db.table.id", "12345")
scache.AddWithDuration(r, dataID, user)
```

## Read items from the cache

```go
dataID := data.NewID("db.table.id", "12345")
var u User
ok := scache.Get(r, dataID, &u)
```