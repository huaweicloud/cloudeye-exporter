package publicips

import (
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/pagination"
)

type PublicIPRequest struct {
	// Specifies the type of the elastic IP address. The value can the
	// 5_telcom, 5_union, 5_bgp, or 5_sbgp. China Northeast: 5_telcom and 5_union China
	// South: 5_sbgp China East: 5_sbgp China North: 5_bgp and 5_sbgp The value must be a
	// type supported by the system. The value can be 5_telcom, 5_union, or 5_bgp.
	Type string `json:"type" required:"true"`

	// Specifies the elastic IP address to be obtained. The value must
	// be a valid IP address in the available IP address segment.
	IpAddress string `json:"ip_address,omitempty"`

	//Value range: 4, 6, respectively, to create ipv4 and ipv6, when not created ipv4 by default
	IPVersion int `json:"ip_version,omitempty"`
}

type BandWidth struct {
	// Specifies the bandwidth name. The value is a string of 1 to 64
	// characters that can contain letters, digits, underscores (_), and hyphens (-). This
	// parameter is mandatory when share_type is set to PER and is optional when share_type
	// is set to WHOLE with an ID specified.
	Name string `json:"name,omitempty"`

	// Specifies the bandwidth size. The value ranges from 1 Mbit/s to
	// 300 Mbit/s. This parameter is mandatory when share_type is set to PER and is optional
	// when share_type is set to WHOLE with an ID specified.
	Size int `json:"size,omitempty"`

	// Specifies the ID of the bandwidth. You can specify an earlier
	// shared bandwidth when applying for an elastic IP address for the bandwidth whose type
	// is set to WHOLE. The bandwidth whose type is set to WHOLE exclusively uses its own
	// ID. The value can be the ID of the bandwidth whose type is set to WHOLE.
	ID string `json:"id,omitempty"`

	// Specifies whether the bandwidth is shared or exclusive. The
	// value can be PER or WHOLE.
	ShareType string `json:"share_type" required:"true"`

	// Specifies the charging mode (by traffic or by bandwidth). The
	// value can be bandwidth or traffic. If the value is an empty character string or no
	// value is specified, default value bandwidth is used.
	ChargeMode string `json:"charge_mode,omitempty"`
}

type CreateOpts struct {
	// Specifies the elastic IP address objects.
	Publicip PublicIPRequest `json:"publicip" required:"true"`

	// Specifies the bandwidth objects.
	Bandwidth BandWidth `json:"bandwidth" required:"true"`
	//	Enterprise project ID. The maximum length is 36 bytes, with the U-ID format of the hyphen "-", or the string "0".
	//When creating an elastic public IP address, bind the enterprise project ID to the elastic public network IP.
	EnterpriseProjectId string `json:"enterprise_project_id,omitempty"`
}

type CreateOptsBuilder interface {
	ToCreatePublicIPMap() (map[string]interface{}, error)
}

func (opts CreateOpts) ToCreatePublicIPMap() (map[string]interface{}, error) {
	b, err := golangsdk.BuildRequestBody(opts, "")
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Create(client *golangsdk.ServiceClient, opts CreateOptsBuilder) (r CreateResult) {
	b, err := opts.ToCreatePublicIPMap()
	if err != nil {
		r.Err = err
		return
	}

	_, r.Err = client.Post(CreateURL(client), b, &r.Body, &golangsdk.RequestOpts{OkCodes: []int{200}})
	return
}

func Delete(client *golangsdk.ServiceClient, publicipId string) (r DeleteResult) {
	url := DeleteURL(client, publicipId)
	_, r.Err = client.Delete(url, nil)
	return
}

func Get(client *golangsdk.ServiceClient, publicipId string) (r GetResult) {
	url := GetURL(client, publicipId)
	_, r.Err = client.Get(url, &r.Body, &golangsdk.RequestOpts{})
	return
}

type ListOpts struct {
	// Specifies the resource ID of pagination query. If the parameter
	// is left blank, only resources on the first page are queried.
	Marker string `q:"marker"`

	// Specifies the number of records returned on each page. The
	// value ranges from 0 to intmax.
	Limit int `q:"limit"`

	//Value range: 4, 6, respectively, to create ipv4 and ipv6, when not created ipv4 by default
	IPVersion int `q:"ip_version"`

	// enterprise_project_id
	// You can use this field to filter the elastic public IP under an enterprise project.
	EnterpriseProjectId string `q:"enterprise_project_id"`
}

type ListOptsBuilder interface {
	ToListPublicIPQuery() (string, error)
}

func (opts ListOpts) ToListPublicIPQuery() (string, error) {
	q, err := golangsdk.BuildQueryString(opts)
	return q.String(), err
}

func List(client *golangsdk.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	url := ListURL(client)
	if opts != nil {
		query, err := opts.ToListPublicIPQuery()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}

	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return PublicIPPage{pagination.LinkedPageBase{PageResult: r}}
	})
}

type UpdateOpts struct {
	// Specifies the port ID.
	PortId string `json:"port_id,omitempty"`

	//Value range: 4, 6, respectively, to create ipv4 and ipv6, when not created ipv4 by default
	IPVersion int `json:"ip_version,omitempty"`
}

type UpdateOptsBuilder interface {
	ToUpdatePublicIPMap() (map[string]interface{}, error)
}

func (opts UpdateOpts) ToUpdatePublicIPMap() (map[string]interface{}, error) {
	b, err := golangsdk.BuildRequestBody(opts, "publicip")
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Update(client *golangsdk.ServiceClient, publicipId string, opts UpdateOptsBuilder) (r UpdateResult) {
	b, err := opts.ToUpdatePublicIPMap()
	if err != nil {
		r.Err = err
		return
	}

	_, r.Err = client.Put(UpdateURL(client, publicipId), b, &r.Body, &golangsdk.RequestOpts{OkCodes: []int{200}})
	return
}
