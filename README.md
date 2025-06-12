# Go Dcp MongoDB

**Go Dcp MongoDB** streams documents from Couchbase Database Change Protocol (DCP) and writes to
MongoDB document in near real-time.

## Example

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
