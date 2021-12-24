package goconsul

import (
	"fmt"

	"github.com/gogf/gf/text/gstr"
)

type ServiceInstance struct {
	Check *ServiceCheck
	KVCmd
	ServiceName  string
	Host         string
	Port         int
	Metadata     map[string]string
	InstanceId   string
	Tags         []string
	LanIps       []string
	AnnouncedIps []string
	registered   bool
}

func NewInstance(
	serviceName string,
	host string,
	port int,
	metadata map[string]string,
	tags []string,
	instanceId string,
	lanIps []string,
	announcedIps []string,
	client *Client,
	dataRoot string) *ServiceInstance {

	instance := &ServiceInstance{
		registered: false,
	}

	instance.Constructor(
		serviceName,
		host,
		port,
		metadata,
		tags,
		instanceId,
		lanIps,
		announcedIps,
		client,
		dataRoot)

	return instance
}

func (ins *ServiceInstance) Constructor(
	serviceName string,
	host string,
	port int,
	metadata map[string]string,
	tags []string,
	instanceId string,
	lanIps []string,
	announcedIps []string,
	client *Client,
	dataRoot string) {

	ips := ""
	for _, ip := range announcedIps {
		ips += ip + ","
	}

	ips = gstr.TrimRight(ips, ",")
	if metadata == nil {
		metadata = make(map[string]string)
	}

	metadata["AnnouncedIp"] = ips

	ips = ""
	for _, ip := range lanIps {
		ips += ip + ","
	}
	ips = gstr.TrimRight(ips, ",")
	metadata["LanIp"] = ips

	fullTags := []string{
		fmt.Sprintf("serviceName=%s", serviceName),
		fmt.Sprintf("instanceId=%s", instanceId),
	}

	fullTags = append(fullTags, tags...)

	if client == nil {
		client = DefaultClient()
	}

	ins.ServiceName = serviceName
	ins.Host = host
	ins.Port = port
	ins.Metadata = metadata
	ins.Tags = fullTags
	ins.InstanceId = instanceId
	ins.LanIps = lanIps
	ins.AnnouncedIps = announcedIps
	ins.KVCmd = KVCmd{
		client:   client,
		DataRoot: dataRoot,
	}
}

func (ins *ServiceInstance) GetId() string {
	return ins.InstanceId
}

func (ins *ServiceInstance) GetName() string {
	return ins.ServiceName
}

func (ins *ServiceInstance) GetHost() string {
	return ins.Host
}

func (ins *ServiceInstance) GetPort() int {
	return ins.Port
}

func (ins *ServiceInstance) GetMetadata() map[string]string {
	return ins.Metadata
}

func (ins *ServiceInstance) GetInstanceId() string {
	return ins.InstanceId
}

func (ins *ServiceInstance) GetTags() []string {
	return ins.Tags
}

func (ins *ServiceInstance) GetLanIps() []string {
	return ins.LanIps
}

func (ins *ServiceInstance) GetAnnouncedIps() []string {
	return ins.AnnouncedIps
}

func (ins *ServiceInstance) GetMetadataValue(key string) (string, bool) {
	val, found := ins.Metadata[key]

	return val, found
}

func (ins *ServiceInstance) GetServiceCheck() *ServiceCheck {
	return ins.Check
}

func (ins *ServiceInstance) Register() error {
	if ins.client == nil {
		ins.client = DefaultClient()
	}

	if err := ins.client.Register(ins); err != nil {
		return err
	}

	ins.registered = true

	return nil
}

func (ins *ServiceInstance) Deregister() {
	if ins.client == nil {
		ins.client = DefaultClient()
	}

	if ins.registered {
		ins.registered = false
		ins.client.Deregister(ins.InstanceId)
	}
}

func (ins *ServiceInstance) IsRegistered() bool {
	return ins.registered
}
