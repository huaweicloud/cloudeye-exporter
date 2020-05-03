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
	"sys_elb":          []string{"name", "vip_address"},
	"sys_elb_listener": []string{"name", "port"},
	"sys_nat":          []string{"name"},
	"sys_rds":          []string{"name", "role"},
	"sys_dcs":          []string{"name", "engine"},
	"sys_dms":          []string{"name"},
	// Added extenstion labeles name for each service
	// for example:
	// "sys_vpc": []string{"name", "cidr"},
}

func (exporter *BaseHuaweiCloudExporter) getAllResource(namespace string) map[string][]string {
	resourceInfos := map[string][]string{}
	switch namespace {
	case "SYS.ELB":
		allELBs, err := getAllELB(exporter.ClientConfig)
		if err != nil {
			fmt.Println("Get all ELB error:", err.Error())
			break
		}
		for _, elb := range *allELBs {
			values := []string{}
			values = append(values, elb.Name)
			values = append(values, elb.VipAddress)
			resourceInfos[elb.ID] = values
		}

		allListeners, err := getAllListener(exporter.ClientConfig)
		if err != nil {
			fmt.Println("Get all Listener error:", err.Error())
			break
		}
		for _, listener := range *allListeners {
			values := []string{}
			values = append(values, listener.Name)
			values = append(values, fmt.Sprintf("%d", listener.ProtocolPort))
			resourceInfos[listener.ID] = values
		}
	case "SYS.NAT":
		allnat, err := getAllNat(exporter.ClientConfig)
		if err != nil {
			fmt.Println("Get all Nat error:", err.Error())
			break
		}
		for _, nat := range *allnat {
			values := []string{}
			values = append(values, nat.Name)
			resourceInfos[nat.ID] = values
		}
	case "SYS.RDS":
		allrds, err := getAllRds(exporter.ClientConfig)
		if err != nil {
			fmt.Println("Get all Rds error:", err.Error())
			break
		}
		for _, rds := range allrds.Instances {
			for _, node := range rds.Nodes {
				nodes := []string{}
				nodes = append(nodes, node.Name)
				nodes = append(nodes, node.Role)
				resourceInfos[node.Id] = nodes
			}
		}
	case "SYS.DMS":
		alldms, err := getAllDms(exporter.ClientConfig)
		if err != nil {
			fmt.Println("Get all Dms error:", err.Error())
			break
		}
		for _, dms := range alldms.Instances {
			values := []string{}
			values = append(values, dms.Name)
			resourceInfos[dms.InstanceID] = values
		}

		allqueue, err := getAllDmsQueue(exporter.ClientConfig)
		if err != nil {
			fmt.Println("Get all Dms Queue error:", err.Error())
			break
		}
		for _, queue := range *allqueue {
			values := []string{}
			values = append(values, queue.Name)
			resourceInfos[queue.ID] = values
		}
	case "SYS.DCS":
		alldcs, err := getAllDcs(exporter.ClientConfig)
		if err != nil {
			fmt.Println("Get all Dcs error:", err.Error())
			break
		}
		for _, dcs := range alldcs.Instances {
			values := []string{}
			values = append(values, dcs.Name)
			values = append(values, dcs.Engine)
			resourceInfos[dcs.InstanceID] = values
		}
		// added another resource extenstion labels and label values.
		// for example:
		// case "SYS.OBS":
		//    allvpc := getAllOBS(exporter.ClientConfig)
		//
	}

	return resourceInfos
}
