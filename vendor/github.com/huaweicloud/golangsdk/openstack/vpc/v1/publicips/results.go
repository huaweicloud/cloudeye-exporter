package publicips

import (
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/pagination"
)

type PublicIP struct {
	// Specifies the ID of the elastic IP address, which uniquely
	// identifies the elastic IP address.
	ID string `json:"id"`

	// Specifies the status of the elastic IP address.
	Status string `json:"status"`

	// ext data for order info
	Profile Profile `json:"profile"`

	// Specifies the type of the elastic IP address.
	Type string `json:"type"`

	// Specifies the obtained elastic IP address.
	PublicIpAddress string `json:"public_ip_address"`

	// Specifies the obtained elastic IPv6 address.
	PublicIpV6Address string `json:"public_ipv6_address"`

	//Value range: 4, 6, respectively, to create ipv4 and ipv6, when not created ipv4 by default
	IPVersion int `json:"ip_version"`

	// Specifies the private IP address bound to the elastic IP
	// address.
	PrivateIpAddress string `json:"private_ip_address,omitempty"`

	// Specifies the port ID.
	PortId string `json:"port_id,omitempty"`

	// Specifies the tenant ID of the operator.
	TenantId string `json:"tenant_id"`

	// Specifies the time for applying for the elastic IP address.
	CreateTime string `json:"create_time"`

	// Specifies the bandwidth ID of the elastic IP address.
	BandwidthId string `json:"bandwidth_id"`

	// Specifies the bandwidth size.
	BandwidthSize int `json:"bandwidth_size"`

	// Specifies whether the bandwidth is shared or exclusive.
	BandwidthShareType string `json:"bandwidth_share_type"`
	//
	// Specifies the bandwidth name.
	BandwidthName string `json:"bandwidth_name"`

	//	Enterprise project ID. The maximum length is 36 bytes, with the U-ID format of the hyphen "-", or the string "0".
	//When creating an elastic public IP address, bind the enterprise project ID to the elastic public network IP.
	EnterpriseProjectId string `json:"enterprise_project_id,omitempty"`
}

type Profile struct {
	UserID    string `json:"user_id"`
	ProductID string `json:"product_id"`
	RegionID  string `json:"region_id"`
	OrderID   string `json:"order_id"`
}

type PublicIPCreateResp struct {
	// Specifies the ID of the elastic IP address, which uniquely
	// identifies the elastic IP address.
	ID string `json:"id"`

	// Specifies the status of the elastic IP address.
	Status string `json:"status"`

	// Specifies the type of the elastic IP address.
	Type string `json:"type"`

	// Specifies the obtained elastic IP address.
	PublicIpAddress string `json:"public_ip_address"`

	// Specifies the obtained elastic IPv6 address.
	PublicIpV6Address string `json:"public_ipv6_address"`

	//Value range: 4, 6, respectively, to create ipv4 and ipv6, when not created ipv4 by default
	IPVersion int `json:"ip_version"`

	// Specifies the tenant ID of the operator.
	TenantId string `json:"tenant_id"`

	// Specifies the time for applying for the elastic IP address.
	CreateTime string `json:"create_time"`

	// Specifies the bandwidth size.
	BandwidthSize int `json:"bandwidth_size"`

	//	Enterprise project ID. The maximum length is 36 bytes, with the U-ID format of the hyphen "-", or the string "0".
	//When creating an elastic public IP address, bind the enterprise project ID to the elastic public network IP.
	EnterpriseProjectId string `json:"enterprise_project_id,omitempty"`
}

type PublicIPUpdateResp struct {
	// Specifies the ID of the elastic IP address, which uniquely
	// identifies the elastic IP address.
	ID string `json:"id"`

	// Specifies the status of the elastic IP address.
	Status string `json:"status"`

	// Specifies the type of the elastic IP address.
	Type string `json:"type"`

	// Specifies the obtained elastic IP address.
	PublicIpAddress string `json:"public_ip_address"`

	// Specifies the private IP address bound to the elastic IP address.
	PrivateIpAddress string `json:"private_ip_address"`

	// Specifies the obtained elastic IPv6 address.
	PublicIpV6Address string `json:"public_ipv6_address"`

	//Value range: 4, 6, respectively, to create ipv4 and ipv6, when not created ipv4 by default
	IPVersion int `json:"ip_version"`

	// Specifies the port ID.
	PortId string `json:"port_id,omitempty"`

	// Specifies the tenant ID of the operator.
	TenantId string `json:"tenant_id"`

	// Specifies the time for applying for the elastic IP address.
	CreateTime string `json:"create_time"`

	// Specifies the bandwidth size.
	BandwidthSize int `json:"bandwidth_size"`

	// Specifies the bandwidth ID of the elastic IP address.
	BandwidthId string `json:"bandwidth_id"`

	// Specifies whether the bandwidth is shared or exclusive.
	BandwidthShareType string `json:"bandwidth_share_type"`

	// Specifies the bandwidth name.
	BandwidthName string `json:"bandwidth_name"`

	//	Enterprise project ID. The maximum length is 36 bytes, with the U-ID format of the hyphen "-", or the string "0".
	//When creating an elastic public IP address, bind the enterprise project ID to the elastic public network IP.
	EnterpriseProjectId string `json:"enterprise_project_id"`
}
type commonResult struct {
	golangsdk.Result
}

type CreateResult struct {
	commonResult
}

func (r CreateResult) Extract() (*PublicIPCreateResp, error) {
	var entity PublicIPCreateResp
	err := r.ExtractIntoStructPtr(&entity, "publicip")
	return &entity, err
}

type DeleteResult struct {
	golangsdk.ErrResult
}

type GetResult struct {
	commonResult
}

func (r GetResult) Extract() (*PublicIP, error) {
	var entity PublicIP
	err := r.ExtractIntoStructPtr(&entity, "publicip")
	return &entity, err
}

type PublicIPPage struct {
	pagination.LinkedPageBase
}

func (r PublicIPPage) NextPageURL() (string, error) {
	publicIps, err := ExtractPublicIPs(r)
	if err != nil {
		return "", err
	}
	return r.WrapNextPageURL(publicIps[len(publicIps)-1].ID)
}

func ExtractPublicIPs(r pagination.Page) ([]PublicIP, error) {
	var s struct {
		PublicIPs []PublicIP `json:"publicips"`
	}
	err := r.(PublicIPPage).ExtractInto(&s)
	return s.PublicIPs, err
}

// IsEmpty checks whether a NetworkPage struct is empty.
func (r PublicIPPage) IsEmpty() (bool, error) {
	s, err := ExtractPublicIPs(r)
	return len(s) == 0, err
}

type UpdateResult struct {
	commonResult
}

func (r UpdateResult) Extract() (*PublicIPUpdateResp, error) {
	var entity PublicIPUpdateResp
	err := r.ExtractIntoStructPtr(&entity, "publicip")
	return &entity, err
}
