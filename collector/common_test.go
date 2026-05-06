package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock functions for testing
func mockCloudConf() {
	CloudConf = CloudConfig{
		Global: Global{
			HttpSchema: "http",
			HttpHost:   "proxy.example.com",
			HttpPort:   8080,
			UserName:   "user",
			Password:   "pass",
		},
	}
}

func TestGetHttpConfig(t *testing.T) {
	// Set up mock CloudConf
	mockCloudConf()

	// Call the method
	httpConfig := GetHttpConfig()

	// Check the results
	assert.NotNil(t, httpConfig.HttpProxy)
	assert.Equal(t, "http", httpConfig.HttpProxy.Schema)
	assert.Equal(t, "proxy.example.com", httpConfig.HttpProxy.Host)
	assert.Equal(t, 8080, httpConfig.HttpProxy.Port)
	assert.Equal(t, "user", httpConfig.HttpProxy.Username)
	assert.Equal(t, "pass", httpConfig.HttpProxy.Password)
}

func TestGetHttpConfigWithoutProxy(t *testing.T) {
	// Set up mock CloudConf with empty proxy info
	CloudConf = CloudConfig{
		Global: Global{
			HttpSchema: "",
			HttpHost:   "",
			HttpPort:   0,
			UserName:   "user",
			Password:   "pass",
		},
	}

	// Call the method
	httpConfig := GetHttpConfig()

	// Check the results
	assert.Nil(t, httpConfig.HttpProxy)
}

func TestGetHttpConfigWithoutUserInfo(t *testing.T) {
	// Set up mock CloudConf with empty user info
	CloudConf = CloudConfig{
		Global: Global{
			HttpSchema: "http",
			HttpHost:   "proxy.example.com",
			HttpPort:   8080,
			UserName:   "",
			Password:   "",
		},
	}

	// Call the method
	httpConfig := GetHttpConfig()

	// Check the results
	assert.NotNil(t, httpConfig.HttpProxy)
	assert.Equal(t, "http", httpConfig.HttpProxy.Schema)
	assert.Equal(t, "proxy.example.com", httpConfig.HttpProxy.Host)
	assert.Equal(t, 8080, httpConfig.HttpProxy.Port)
	assert.Equal(t, "", httpConfig.HttpProxy.Username)
	assert.Equal(t, "", httpConfig.HttpProxy.Password)
}

func TestIsProxyValid(t *testing.T) {
	// Set up mock CloudConf with valid proxy info
	CloudConf = CloudConfig{
		Global: Global{
			HttpSchema: "http",
			HttpHost:   "proxy.example.com",
			HttpPort:   8080,
		},
	}
	assert.True(t, isProxyValid())

	// Set up mock CloudConf with invalid proxy info
	CloudConf = CloudConfig{
		Global: Global{
			HttpSchema: "",
			HttpHost:   "",
			HttpPort:   0,
		},
	}
	assert.False(t, isProxyValid())
}

func TestIsUserInfoValid(t *testing.T) {
	// Set up mock CloudConf with valid user info
	CloudConf = CloudConfig{
		Global: Global{
			UserName: "user",
			Password: "pass",
		},
	}
	assert.True(t, isUserInfoValid())

	// Set up mock CloudConf with invalid user info
	CloudConf = CloudConfig{
		Global: Global{
			UserName: "",
			Password: "",
		},
	}
	assert.False(t, isUserInfoValid())
}
