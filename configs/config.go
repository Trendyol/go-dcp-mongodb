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
}
