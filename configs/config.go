package config

import (
	"time"

	"github.com/Trendyol/go-dcp/helpers"

	"github.com/Trendyol/go-dcp/config"
)

type Config struct {
	MongoDB MongoDB    `yaml:"mongodb" mapstructure:"mongodb"`
	Dcp     config.Dcp `yaml:",inline" mapstructure:",squash"`
}

type MongoDB struct {
	URI                       string         `yaml:"uri"`
	Username                  string         `yaml:"username"`
	Password                  string         `yaml:"password"`
	Database                  string         `yaml:"database"`
	Collection                string         `yaml:"collection"`
	BatchSizeLimit            int            `yaml:"batchSizeLimit"`
	BatchByteSizeLimit        any            `yaml:"batchByteSizeLimit"`
	ConcurrentRequest         int            `yaml:"concurrentRequest"`
	BatchTickerDuration       time.Duration  `yaml:"batchTickerDuration"`
	BatchCommitTickerDuration *time.Duration `yaml:"batchCommitTickerDuration"`
	ShardKeys                 []string       `yaml:"shardKeys,omitempty" mapstructure:"shardKeys"`

	// Connection Pool Settings
	MaxPoolSize   uint64 `yaml:"maxPoolSize"`
	MinPoolSize   uint64 `yaml:"minPoolSize"`
	MaxIdleTimeMS uint64 `yaml:"maxIdleTimeMS"`

	// Timeout Settings
	ConnectTimeoutMS         uint64 `yaml:"connectTimeoutMS"`
	ServerSelectionTimeoutMS uint64 `yaml:"serverSelectionTimeoutMS"`
	SocketTimeoutMS          uint64 `yaml:"socketTimeoutMS"`
}

func (c *Config) ApplyDefaults() {
	if c.MongoDB.BatchTickerDuration == 0 {
		c.MongoDB.BatchTickerDuration = 10 * time.Second
	}

	if c.MongoDB.BatchSizeLimit == 0 {
		c.MongoDB.BatchSizeLimit = 1000
	}

	if c.MongoDB.BatchByteSizeLimit == nil {
		c.MongoDB.BatchByteSizeLimit = helpers.ResolveUnionIntOrStringValue("10mb")
	}

	if c.MongoDB.ConcurrentRequest == 0 {
		c.MongoDB.ConcurrentRequest = 1
	}

	// Connection Pool Defaults
	if c.MongoDB.MaxPoolSize == 0 {
		c.MongoDB.MaxPoolSize = 100
	}

	if c.MongoDB.MinPoolSize == 0 {
		c.MongoDB.MinPoolSize = 5
	}

	if c.MongoDB.MaxIdleTimeMS == 0 {
		c.MongoDB.MaxIdleTimeMS = 300000 // 5 minutes
	}

	// Timeout Defaults
	if c.MongoDB.ConnectTimeoutMS == 0 {
		c.MongoDB.ConnectTimeoutMS = 10000 // 10 seconds
	}

	if c.MongoDB.ServerSelectionTimeoutMS == 0 {
		c.MongoDB.ServerSelectionTimeoutMS = 30000 // 30 seconds
	}

	if c.MongoDB.SocketTimeoutMS == 0 {
		c.MongoDB.SocketTimeoutMS = 30000 // 30 seconds
	}
}
