# Go Dcp MongoDB

**Go Dcp MongoDB** streams documents from Couchbase Database Change Protocol (DCP) and writes to
MongoDB document in near real-time.

## Features

* **Custom routing** support(see [Example](#example)).
* **Update multiple documents** for a DCP event(see [Example](#example)).
* Handling different DCP events such as **expiration, deletion and mutation**(see [Example](#example)).
* **Managing batch configurations** such as maximum batch size, batch bytes, batch ticker durations.
* **Advanced connection pool management** with configurable pool sizes and idle timeouts.
* **Comprehensive timeout configurations** for connection, server selection, and socket operations.
* **Scale up and down** by custom membership algorithms(Couchbase, KubernetesHa, Kubernetes StatefulSet or
  Static, see [examples](https://github.com/Trendyol/go-dcp#examples)).
* **Easily manageable configurations**.

## Example

[Struct Config](example/struct-config/main.go)

[File Config](example/simple/main.go)

[Default Mapper](example/default-mapper/main.go)

## Configuration

### Dcp Configuration

Check out on [go-dcp](https://github.com/Trendyol/go-dcp#configuration)

### MongoDB Specific Configuration

MongoDB configuration is organized into logical groups for better management:

#### Connection Settings (`mongodb.connection`)

| Variable                      | Type   | Required | Default | Description                                                                                  |
|-------------------------------|--------|----------|---------|----------------------------------------------------------------------------------------------|
| `mongodb.connection.uri`      | string | yes      |         | MongoDB connection URI (e.g., "localhost:27017")                                             |
| `mongodb.connection.database` | string | yes      |         | MongoDB database name                                                                        |
| `mongodb.connection.username` | string | no       |         | MongoDB username for authentication                                                          |
| `mongodb.connection.password` | string | no       |         | MongoDB password for authentication                                                          |

#### Batch Processing Settings (`mongodb.batch`)

| Variable                                | Type          | Required | Default | Description                                                                                                                                                 |
|-----------------------------------------|---------------|----------|---------|-------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `mongodb.batch.sizeLimit`               | int           | no       | 1000    | Maximum message count for batch, if exceed flush will be triggered                                                                                          |
| `mongodb.batch.byteSizeLimit`           | int, string   | no       | 10mb    | Maximum size(byte) for batch, if exceed flush will be triggered. Supports units like "10mb", "1gb"                                                          |
| `mongodb.batch.tickerDuration`          | time.Duration | no       | 10s     | Batch is being flushed automatically at specific time intervals for long waiting messages in batch                                                          |
| `mongodb.batch.commitTickerDuration`    | time.Duration | no       | 0s      | Configures checkpoint offset save time, By default, after batch flushing, the offsets are updated immediately, this period can be increased for performance |
| `mongodb.batch.concurrentRequest`       | int           | no       | 1       | Concurrent bulk request count                                                                                                                               |

#### Connection Pool Settings (`mongodb.connectionPool`)

| Variable                              | Type   | Required | Default | Description                                                                                    |
|---------------------------------------|--------|----------|---------|------------------------------------------------------------------------------------------------|
| `mongodb.connectionPool.maxPoolSize`  | uint64 | no       | 100     | Maximum number of connections in the connection pool                                           |
| `mongodb.connectionPool.minPoolSize`  | uint64 | no       | 5       | Minimum number of connections to maintain in the connection pool                               |
| `mongodb.connectionPool.maxIdleTimeMS`| int64  | no       | 300000  | Maximum time (in milliseconds) a connection can remain idle before being closed (5 minutes)    |

#### Timeout Settings (`mongodb.timeouts`)

| Variable                                    | Type  | Required | Default | Description                                                                                   |
|---------------------------------------------|-------|----------|---------|-----------------------------------------------------------------------------------------------|
| `mongodb.timeouts.connectTimeoutMS`         | int64 | no       | 10000   | Connection timeout in milliseconds (10 seconds)                                               |
| `mongodb.timeouts.serverSelectionTimeoutMS` | int64 | no       | 30000   | Server selection timeout in milliseconds (30 seconds)                                         |
| `mongodb.timeouts.socketTimeoutMS`          | int64 | no       | 30000   | Socket timeout in milliseconds (30 seconds)                                                   |

#### General Settings

| Variable                | Type     | Required | Default | Description                                                                                   |
|-------------------------|----------|----------|---------|-----------------------------------------------------------------------------------------------|
| `mongodb.collection`    | string   | yes      |         | MongoDB collection name                                                                       |
| `mongodb.shardKeys`     | []string | no       |         | List of shard key paths from document for MongoDB sharded clusters. Used in query filters     |

### Configuration Example

```yaml
mongodb:
  connection:
    uri: "localhost:27017"
    database: "exampleDB"
    username: "user"
    password: "pass"
  collection: "exampleCollection"
  batch:
    sizeLimit: 1000
    byteSizeLimit: "10mb"
    tickerDuration: 10s
    concurrentRequest: 2
  connectionPool:
    maxPoolSize: 100
    minPoolSize: 5
    maxIdleTimeMS: 300000  # 5 minutes
  timeouts:
    connectTimeoutMS: 10000  # 10 seconds
    serverSelectionTimeoutMS: 30000  # 30 seconds
    socketTimeoutMS: 30000  # 30 seconds
  shardKeys:
    - "customer.id"
    - "tenant.id"
```

## Exposed metrics

| Metric Name                                                      | Description                    | Labels                                                                                                                                                                              | Value Type |
|------------------------------------------------------------------|--------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------|
| cbgo_mongodb_connector_latency_ms_current                        | Time to adding to the batch.   | N/A                                                                                                                                                                                 | Gauge      |
| cbgo_mongodb_connector_bulk_request_process_latency_ms_current   | Time to process bulk request.  | N/A                                                                                                                                                                                 | Gauge      |
| cbgo_mongodb_connector_update_operations_total                   | Count of update operations     | `collection`: MongoDB collection name, `status`: Operation result (`success`, `error`)                                                                                              | Counter    |
| cbgo_mongodb_connector_delete_operations_total                   | Count of delete operations     | `collection`: MongoDB collection name, `status`: Operation result (`success`, `error`)                                                                                              | Counter    |


You can also use all DCP-related metrics explained [here](https://github.com/Trendyol/go-dcp#exposed-metrics).
All DCP-related metrics are automatically injected. It means you don't need to do anything.

## Grafana Metric Dashboard

[Grafana & Prometheus Example](example/grafana)

## Contributing

Go Dcp MongoDB is always open for direct contributions. For more information please check
our [Contribution Guideline document](./CONTRIBUTING.md).

## License

Released under the [MIT License](LICENSE).