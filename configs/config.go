package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/Trendyol/go-dcp/helpers"

	"github.com/Trendyol/go-dcp/config"
)

type Config struct {
	MongoDB MongoDB    `yaml:"mongodb" mapstructure:"mongodb"`
	Dcp     config.Dcp `yaml:",inline" mapstructure:",squash"`
}

type MongoDB struct {
	Connection     Connection     `yaml:"connection" mapstructure:"connection"`
	Collection     string         `yaml:"collection"`
	Batch          BatchConfig    `yaml:"batch" mapstructure:"batch"`
	ConnectionPool ConnectionPool `yaml:"connectionPool" mapstructure:"connectionPool"`
	Timeouts       Timeouts       `yaml:"timeouts" mapstructure:"timeouts"`
	ShardKeys      []string       `yaml:"shardKeys,omitempty" mapstructure:"shardKeys"`
}

type Connection struct {
	URI      string `yaml:"uri"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type BatchConfig struct {
	SizeLimit            int            `yaml:"sizeLimit"`
	ByteSizeLimit        any            `yaml:"byteSizeLimit"`
	ConcurrentRequest    int            `yaml:"concurrentRequest"`
	TickerDuration       time.Duration  `yaml:"tickerDuration"`
	CommitTickerDuration *time.Duration `yaml:"commitTickerDuration"`
}

type ConnectionPool struct {
	MaxPoolSize   uint64 `yaml:"maxPoolSize"`
	MinPoolSize   uint64 `yaml:"minPoolSize"`
	MaxIdleTimeMS int64  `yaml:"maxIdleTimeMS"`
}

type Timeouts struct {
	ConnectTimeoutMS         int64 `yaml:"connectTimeoutMS"`
	ServerSelectionTimeoutMS int64 `yaml:"serverSelectionTimeoutMS"`
	SocketTimeoutMS          int64 `yaml:"socketTimeoutMS"`
}

func (c *Config) ApplyDefaults() {
	if c.MongoDB.Batch.TickerDuration == 0 {
		c.MongoDB.Batch.TickerDuration = 10 * time.Second
	}

	if c.MongoDB.Batch.SizeLimit == 0 {
		c.MongoDB.Batch.SizeLimit = 1000
	}

	if c.MongoDB.Batch.ByteSizeLimit == nil {
		c.MongoDB.Batch.ByteSizeLimit = helpers.ResolveUnionIntOrStringValue("10mb")
	}

	if c.MongoDB.Batch.ConcurrentRequest == 0 {
		c.MongoDB.Batch.ConcurrentRequest = 1
	}

	if c.MongoDB.ConnectionPool.MaxPoolSize == 0 {
		c.MongoDB.ConnectionPool.MaxPoolSize = 100
	}

	if c.MongoDB.ConnectionPool.MinPoolSize == 0 {
		c.MongoDB.ConnectionPool.MinPoolSize = 5
	}

	if c.MongoDB.ConnectionPool.MaxIdleTimeMS == 0 {
		c.MongoDB.ConnectionPool.MaxIdleTimeMS = 300000 // 5 minutes
	}

	if c.MongoDB.Timeouts.ConnectTimeoutMS == 0 {
		c.MongoDB.Timeouts.ConnectTimeoutMS = 10000 // 10 seconds
	}

	if c.MongoDB.Timeouts.ServerSelectionTimeoutMS == 0 {
		c.MongoDB.Timeouts.ServerSelectionTimeoutMS = 30000 // 30 seconds
	}

	if c.MongoDB.Timeouts.SocketTimeoutMS == 0 {
		c.MongoDB.Timeouts.SocketTimeoutMS = 30000 // 30 seconds
	}
}

func (c *Config) Validate() error {
	if err := c.MongoDB.Validate(); err != nil {
		return fmt.Errorf("mongodb config validation failed: %w", err)
	}
	return nil
}

func (m *MongoDB) Validate() error {
	if err := m.Connection.Validate(); err != nil {
		return fmt.Errorf("connection validation failed: %w", err)
	}

	if err := m.ConnectionPool.Validate(); err != nil {
		return fmt.Errorf("connection pool validation failed: %w", err)
	}

	if isEmpty(m.Collection) {
		return fmt.Errorf("collection is required")
	}

	return nil
}

func (c *Connection) Validate() error {
	if isEmpty(c.URI) {
		return fmt.Errorf("uri is required")
	}

	if isEmpty(c.Database) {
		return fmt.Errorf("database is required")
	}

	if (isNotEmpty(c.Username) && isEmpty(c.Password)) || (isEmpty(c.Username) && isNotEmpty(c.Password)) {
		return fmt.Errorf("username and password must be provided together")
	}

	return nil
}

func (cp *ConnectionPool) Validate() error {
	if cp.MinPoolSize > cp.MaxPoolSize {
		return fmt.Errorf("minPoolSize (%d) cannot be greater than maxPoolSize (%d)",
			cp.MinPoolSize, cp.MaxPoolSize)
	}

	return nil
}

func isEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

func isNotEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}
