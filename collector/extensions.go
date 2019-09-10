// Copyright 2019 HuaweiCloud.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"fmt"
)

// If the extension labels have to added in this exporter, you only have
// to add the code to the following two parts.
// 1. Added the new labels name to defaultExtensionLabels
// 2. Added the new labels values to getAllResource
var defaultExtensionLabels = map[string][]string{
	"sys_elb": []string{"name", "vip_address"},
	"sys_elb_listener": []string{"name", "port"},
	// Added extenstion labeles name for each service
	// for example:
	// "sys_vpc": []string{"name", "cidr"},
}


func (exporter *BaseHuaweiCloudExporter) getAllResource(namespace string) (map[string][]string) {
	resourceInfos := map[string][]string{}
	switch namespace {
	case "SYS.ELB":
		allELBs, _ := getAllELB(exporter.ClientConfig)
		for _, elb := range *allELBs {
			values := []string{}
			values = append(values, elb.Name)
			values = append(values, elb.VipAddress)
			resourceInfos[elb.ID] = values
		}

		allListeners, _ := getAllListener(exporter.ClientConfig)
		for _, listener := range *allListeners {
			values := []string{}
			values = append(values, listener.Name)
			values = append(values, fmt.Sprintf("%d", listener.ProtocolPort))
			resourceInfos[listener.ID] = values
		}
	// added another resource extenstion labels and label values.
	// for example:
	// case "SYS.VPC":
	//    allvpc := getAllVPC(exporter.ClientConfig)
	//
	}

	return resourceInfos
}
