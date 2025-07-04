# Go Dcp MongoDB

**Go Dcp MongoDB** streams documents from Couchbase Database Change Protocol (DCP) and writes to
MongoDB document in near real-time.

## Features

* **Custom routing** support(see [Example](#example)).
* **Update multiple documents** for a DCP event(see [Example](#example)).
* Handling different DCP events such as **expiration, deletion and mutation**(see [Example](#example)).
* **Managing batch configurations** such as maximum batch size, batch bytes, batch ticker durations.
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

| Variable                            | Type              | Required | Default | Description                                                                                                                                                  |                                                           
|-------------------------------------|-------------------|----------|---------|--------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `mongodb.uri`                       | string            | yes      |         | Defines which Couchbase collection events will be written to which collection.                                                                               |
| `mongodb.database`                  | string            | yes      |         | Defines MongoDB database name.                                                                                                                               |
| `mongodb.collection`                | string            | yes      |         | Defines MongoDB collection name.                                                                                                                             |
| `mongodb.username`                  | string            | no       |         | The username of MongoDB.                                                                                                                                     |
| `mongodb.password`                  | string            | no       |         | The password of MongoDB.                                                                                                                                     |                                                                                                                          |
| `mongodb.batchSizeLimit`            | int               | no       | 1000    | Maximum message count for batch, if exceed flush will be triggered.                                                                                          |
| `mongodb.batchTickerDuration`       | time.Duration     | no       | 10s     | Batch is being flushed automatically at specific time intervals for long waiting messages in batch.                                                          |
| `mongodb.batchCommitTickerDuration` | time.Duration     | no       | 0s      | Configures checkpoint offset save time, By default, after batch flushing, the offsets are updated immediately, this period can be increased for performance. |
| `mongodb.batchByteSizeLimit`        | int, string       | no       | 10mb    | Maximum size(byte) for batch, if exceed flush will be triggered. `10mb` is default.                                                                          |
| `mongodb.concurrentRequest`         | int               | no       | 1       | Concurrent bulk request count.                                                                                                                               |
| `mongodb.shardKeys`                 | []string          | no       |         | List of shard key paths from document for MongoDB sharded clusters. Used in query filters.                                                                   |

## Exposed metrics

| Metric Name                                                      | Description                   | Labels                                                                                                                                                                             | Value Type |
|------------------------------------------------------------------|-------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------|
| cbgo_mongodb_connector_latency_ms_current                        | Time to adding to the batch.  | N/A                                                                                                                                                                                | Gauge      |
| cbgo_mongodb_connector_bulk_request_process_latency_ms_current   | Time to process bulk request. | N/A                                                                                                                                                                                | Gauge      |
| cbgo_mongodb_connector_action_total_current                      | Count mongodb actions         | `action_type`: Type of action (e.g., `delete`) `result`: Result of the action (e.g., `success`, `error`)  `database_name`: The name of the database to which the action is applied | Counter    |

You can also use all DCP-related metrics explained [here](https://github.com/Trendyol/go-dcp#exposed-metrics).
All DCP-related metrics are automatically injected. It means you don't need to do anything.

## Grafana Metric Dashboard

[Grafana & Prometheus Example](example/grafana)

## Contributing

Go Dcp MongoDB is always open for direct contributions. For more information please check
our [Contribution Guideline document](./CONTRIBUTING.md).

## License

Released under the [MIT License](LICENSE).