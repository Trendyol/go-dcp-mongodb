package config

import (
	"testing"
	"time"

	"github.com/Trendyol/go-dcp/helpers"
	"github.com/stretchr/testify/assert"
)

func TestConfig_ApplyDefaults(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		expected *Config
	}{
		{
			name: "should apply all default values",
			config: &Config{
				MongoDB: MongoDB{},
			},
			expected: &Config{
				MongoDB: MongoDB{
					Batch: BatchConfig{
						TickerDuration:    10 * time.Second,
						SizeLimit:         1000,
						ByteSizeLimit:     helpers.ResolveUnionIntOrStringValue("10mb"),
						ConcurrentRequest: 1,
					},
					ConnectionPool: ConnectionPool{
						MaxPoolSize:   100,
						MinPoolSize:   5,
						MaxIdleTimeMS: 300000,
					},
					Timeouts: Timeouts{
						ConnectTimeoutMS:         10000,
						ServerSelectionTimeoutMS: 30000,
						SocketTimeoutMS:          30000,
					},
				},
			},
		},
		{
			name: "should preserve existing values",
			config: &Config{
				MongoDB: MongoDB{
					Batch: BatchConfig{
						TickerDuration:    5 * time.Second,
						SizeLimit:         500,
						ConcurrentRequest: 2,
					},
					ConnectionPool: ConnectionPool{
						MaxPoolSize:   50,
						MinPoolSize:   10,
						MaxIdleTimeMS: 600000,
					},
					Timeouts: Timeouts{
						ConnectTimeoutMS:         5000,
						ServerSelectionTimeoutMS: 15000,
						SocketTimeoutMS:          15000,
					},
				},
			},
			expected: &Config{
				MongoDB: MongoDB{
					Batch: BatchConfig{
						TickerDuration:    5 * time.Second,
						SizeLimit:         500,
						ByteSizeLimit:     helpers.ResolveUnionIntOrStringValue("10mb"),
						ConcurrentRequest: 2,
					},
					ConnectionPool: ConnectionPool{
						MaxPoolSize:   50,
						MinPoolSize:   10,
						MaxIdleTimeMS: 600000,
					},
					Timeouts: Timeouts{
						ConnectTimeoutMS:         5000,
						ServerSelectionTimeoutMS: 15000,
						SocketTimeoutMS:          15000,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.ApplyDefaults()
			assert.Equal(t, tt.expected.MongoDB.Batch.TickerDuration, tt.config.MongoDB.Batch.TickerDuration)
			assert.Equal(t, tt.expected.MongoDB.Batch.SizeLimit, tt.config.MongoDB.Batch.SizeLimit)
			assert.Equal(t, tt.expected.MongoDB.Batch.ConcurrentRequest, tt.config.MongoDB.Batch.ConcurrentRequest)
			assert.Equal(t, tt.expected.MongoDB.ConnectionPool.MaxPoolSize, tt.config.MongoDB.ConnectionPool.MaxPoolSize)
			assert.Equal(t, tt.expected.MongoDB.ConnectionPool.MinPoolSize, tt.config.MongoDB.ConnectionPool.MinPoolSize)
			assert.Equal(t, tt.expected.MongoDB.ConnectionPool.MaxIdleTimeMS, tt.config.MongoDB.ConnectionPool.MaxIdleTimeMS)
			assert.Equal(t, tt.expected.MongoDB.Timeouts.ConnectTimeoutMS, tt.config.MongoDB.Timeouts.ConnectTimeoutMS)
			assert.Equal(t, tt.expected.MongoDB.Timeouts.ServerSelectionTimeoutMS, tt.config.MongoDB.Timeouts.ServerSelectionTimeoutMS)
			assert.Equal(t, tt.expected.MongoDB.Timeouts.SocketTimeoutMS, tt.config.MongoDB.Timeouts.SocketTimeoutMS)
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				MongoDB: MongoDB{
					Connection: Connection{
						URI:      "mongodb://localhost:27017",
						Database: "testdb",
					},
					CollectionMapping: map[string]string{
						"_default": "testcollection",
					},
					ConnectionPool: ConnectionPool{
						MaxPoolSize: 100,
						MinPoolSize: 5,
					},
				},
			},
			expectErr: false,
		},
		{
			name: "invalid mongodb config",
			config: &Config{
				MongoDB: MongoDB{
					Connection: Connection{
						URI: "",
					},
				},
			},
			expectErr: true,
			errMsg:    "mongodb config validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMongoDB_Validate(t *testing.T) {
	tests := []struct {
		name      string
		mongodb   *MongoDB
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid mongodb config",
			mongodb: &MongoDB{
				Connection: Connection{
					URI:      "mongodb://localhost:27017",
					Database: "testdb",
				},
				CollectionMapping: map[string]string{
					"_default": "testcollection",
				},
				ConnectionPool: ConnectionPool{
					MaxPoolSize: 100,
					MinPoolSize: 5,
				},
			},
			expectErr: false,
		},
		{
			name: "invalid connection",
			mongodb: &MongoDB{
				Connection: Connection{
					URI: "",
				},
				CollectionMapping: map[string]string{
					"_default": "testcollection",
				},
			},
			expectErr: true,
			errMsg:    "connection validation failed",
		},
		{
			name: "invalid connection pool",
			mongodb: &MongoDB{
				Connection: Connection{
					URI:      "mongodb://localhost:27017",
					Database: "testdb",
				},
				CollectionMapping: map[string]string{
					"_default": "testcollection",
				},
				ConnectionPool: ConnectionPool{
					MaxPoolSize: 5,
					MinPoolSize: 10,
				},
			},
			expectErr: true,
			errMsg:    "connection pool validation failed",
		},
		{
			name: "empty collection mapping",
			mongodb: &MongoDB{
				Connection: Connection{
					URI:      "mongodb://localhost:27017",
					Database: "testdb",
				},
				CollectionMapping: map[string]string{},
				ConnectionPool: ConnectionPool{
					MaxPoolSize: 100,
					MinPoolSize: 5,
				},
			},
			expectErr: true,
			errMsg:    "collectionMapping is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mongodb.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConnection_Validate(t *testing.T) {
	tests := []struct {
		name       string
		connection *Connection
		expectErr  bool
		errMsg     string
	}{
		{
			name: "valid connection",
			connection: &Connection{
				URI:      "mongodb://localhost:27017",
				Database: "testdb",
			},
			expectErr: false,
		},
		{
			name: "valid connection with auth",
			connection: &Connection{
				URI:      "mongodb://localhost:27017",
				Database: "testdb",
				Username: "user",
				Password: "pass",
			},
			expectErr: false,
		},
		{
			name: "empty URI",
			connection: &Connection{
				URI:      "",
				Database: "testdb",
			},
			expectErr: true,
			errMsg:    "uri is required",
		},
		{
			name: "empty database",
			connection: &Connection{
				URI:      "mongodb://localhost:27017",
				Database: "",
			},
			expectErr: true,
			errMsg:    "database is required",
		},
		{
			name: "only username provided",
			connection: &Connection{
				URI:      "mongodb://localhost:27017",
				Database: "testdb",
				Username: "user",
			},
			expectErr: true,
			errMsg:    "username and password must be provided together",
		},
		{
			name: "only password provided",
			connection: &Connection{
				URI:      "mongodb://localhost:27017",
				Database: "testdb",
				Password: "pass",
			},
			expectErr: true,
			errMsg:    "username and password must be provided together",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.connection.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConnectionPool_Validate(t *testing.T) {
	tests := []struct {
		name           string
		connectionPool *ConnectionPool
		expectErr      bool
		errMsg         string
	}{
		{
			name: "valid connection pool",
			connectionPool: &ConnectionPool{
				MaxPoolSize: 100,
				MinPoolSize: 5,
			},
			expectErr: false,
		},
		{
			name: "equal pool size values",
			connectionPool: &ConnectionPool{
				MaxPoolSize: 50,
				MinPoolSize: 50,
			},
			expectErr: false,
		},
		{
			name: "minPoolSize > maxPoolSize",
			connectionPool: &ConnectionPool{
				MaxPoolSize: 5,
				MinPoolSize: 10,
			},
			expectErr: true,
			errMsg:    "minPoolSize (10) cannot be greater than maxPoolSize (5)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.connectionPool.Validate()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "only whitespace",
			input:    "   ",
			expected: true,
		},
		{
			name:     "tab and whitespace",
			input:    "\t  \n",
			expected: true,
		},
		{
			name:     "non-empty string",
			input:    "test",
			expected: false,
		},
		{
			name:     "string with whitespace",
			input:    "  test  ",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isEmpty(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
