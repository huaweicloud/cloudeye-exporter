package collector

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestName(t *testing.T) {
	err := InitMetricConf()
	assert.Equal(t, true, err != nil)
	metricConf = map[string]MetricConf{
		"SYS.ECS": {
			Resource: "rms",
			DimMetricName: map[string][]string{
				"instance_id": []string{"cpu_util", "mem_util", "disk_util_inband"},
			},
		},
	}
	mconf := getMetricConfigMap("TEST.ECS")
	assert.Equal(t, true, mconf == nil)

	mconf = getMetricConfigMap("SYS.ECS")
	assert.Equal(t, 1, len(mconf))
}

func TestInitConfigSecurityModIsTrue(t *testing.T) {
	SecurityMod = true
	TmpSK = "tmpSK"
	TmpAK = "tmpAK"
	CloudConf.Auth.ProjectID = "testProjectId"
	CloudConf.Auth.ProjectName = "testProjectName"
	err := InitConfig()
	if err != nil {
		return
	}

	assert.Equal(t, TmpAK, conf.AccessKey)
	assert.Equal(t, TmpSK, conf.SecretKey)
}

func TestInitConfigSecurityModIsFalse(t *testing.T) {
	SecurityMod = false
	CloudConf.Auth.AccessKey = "tmpSK"
	CloudConf.Auth.SecretKey = "tmpAK"
	CloudConf.Auth.ProjectID = "testProjectId"
	CloudConf.Auth.ProjectName = "testProjectName"
	err := InitConfig()
	if err != nil {
		return
	}

	assert.Equal(t, CloudConf.Auth.AccessKey, conf.AccessKey)
	assert.Equal(t, CloudConf.Auth.SecretKey, conf.SecretKey)
}
