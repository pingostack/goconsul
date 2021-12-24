package goconsul

import (
	"fmt"

	"github.com/gogf/gf/text/gstr"
	"github.com/hashicorp/consul/api"
)

type ServiceCheck struct {
	api.AgentServiceCheck
}

type IServiceInstance interface {
	GetName() string
	GetHost() string
	GetPort() int
	GetMetadata() map[string]string
	GetId() string
	GetTags() []string
	GetLanIps() []string
	GetAnnouncedIps() []string
	GetServiceCheck() *ServiceCheck
}

func (c *Client) Register(instance *ServiceInstance) error {

	// 创建注册到consul的服务到
	registration := new(api.AgentServiceRegistration)
	registration.ID = instance.GetId()
	registration.Name = instance.GetName()
	registration.Port = instance.GetPort()
	registration.Tags = instance.GetTags()
	registration.Meta = instance.GetMetadata()
	registration.Address = instance.GetHost()
	registration.TaggedAddresses = make(map[string]api.ServiceAddress)
	for k, v := range instance.GetLanIps() {
		registration.TaggedAddresses[fmt.Sprintf("Lan%v", k)] = api.ServiceAddress{
			Address: v,
			Port:    instance.GetPort(),
		}
	}
	for k, v := range instance.GetAnnouncedIps() {
		registration.TaggedAddresses[fmt.Sprintf("Announced%v", k)] = api.ServiceAddress{
			Address: v,
			Port:    instance.GetPort(),
		}
	}
	registration.Check = &instance.GetServiceCheck().AgentServiceCheck

	// 注册服务到consul
	err := c.Agent().ServiceRegister(registration)
	if err != nil {
		fmt.Println(err)
		return err
	}

	c.insMap.Store(instance.GetId(), instance)

	return nil
}

func (c *Client) DeregisterAll() {

	c.insMap.Range(func(key, value interface{}) bool {
		c.Agent().ServiceDeregister(key.(string))
		c.insMap.Delete(key)
		return true
	})
}

func (c *Client) Deregister(id string) {
	c.Agent().ServiceDeregister(id)
	c.insMap.Delete(id)
}

func (c *Client) DiscoverCatalogInstancesWithTags(serviceName string, tags []string) ([]*ServiceInstance, error) {

	catalogService, _, err := c.Catalog().ServiceMultipleTags(serviceName, tags, nil)
	if err != nil {
		return nil, err
	}

	if len(catalogService) == 0 {
		return nil, fmt.Errorf("not found")
	}

	result := make([]*ServiceInstance, len(catalogService))
	for index, server := range catalogService {
		announcedIps := make([]string, 0)
		announcedIpStr, found := server.ServiceMeta["AnnouncedIp"]
		if found && announcedIpStr != "" {
			announcedIps = gstr.Split(announcedIpStr, ",")
		}

		lanIps := make([]string, 0)
		lanIpStr, found := server.ServiceMeta["LanIp"]
		if found && lanIpStr != "" {
			lanIps = gstr.Split(lanIpStr, ",")
		}

		dataRoot, found := server.ServiceMeta["DataRoot"]
		if !found {
			dataRoot = ""
		}

		s := NewInstance(
			server.ServiceName,
			server.Address,
			server.ServicePort,
			server.ServiceMeta,
			server.ServiceTags,
			server.ServiceID,
			lanIps,
			announcedIps,
			c,
			dataRoot,
		)

		result[index] = s
	}

	return result, nil
}

func (c *Client) DiscoverCatalogInstancesWithName(serviceName string) ([]*ServiceInstance, error) {
	return c.DiscoverCatalogInstancesWithTags(serviceName, nil)
}

func (c *Client) DiscoverCatalogInstanceWithId(name string, id string) (*ServiceInstance, error) {
	tags := []string{
		fmt.Sprintf("serviceName:%v", name),
		fmt.Sprintf("instanceId:%v", id),
	}

	ins, err := c.DiscoverCatalogInstancesWithTags(name, tags)
	if err != nil {
		return nil, err
	}

	if len(ins) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return ins[0], nil
}

func (c *Client) DiscoverHealthInstancesWithTags(name string, tags []string, passing bool) ([]*ServiceInstance, error) {

	servers, _, err := c.Health().ServiceMultipleTags(name, tags, passing, nil)
	if err != nil {
		return nil, err
	}

	if len(servers) == 0 {
		return nil, fmt.Errorf("not found")
	}

	result := make([]*ServiceInstance, len(servers))
	for index, server := range servers {
		announcedIps := make([]string, 0)
		announcedIpStr, found := server.Service.Meta["AnnouncedIp"]
		if found && announcedIpStr != "" {
			announcedIps = gstr.Split(announcedIpStr, ",")
		}

		lanIps := make([]string, 0)
		lanIpStr, found := server.Service.Meta["LanIp"]
		if found && lanIpStr != "" {
			lanIps = gstr.Split(lanIpStr, ",")
		}

		dataRoot, found := server.Service.Meta["DataRoot"]
		if !found {
			dataRoot = ""
		}

		s := NewInstance(
			server.Service.Service,
			server.Service.Address,
			server.Service.Port,
			server.Service.Meta,
			server.Service.Tags,
			server.Service.ID,
			lanIps,
			announcedIps,
			c,
			dataRoot,
		)

		result[index] = s
	}

	return result, nil
}

func (c *Client) DiscoverHealthInstanceWithId(name string, id string, passing bool) (*ServiceInstance, error) {
	tags := []string{
		fmt.Sprintf("instanceId:%v", id),
	}

	ins, err := c.DiscoverHealthInstancesWithTags(name, tags, passing)
	if err != nil {
		return nil, err
	}

	if len(ins) == 0 {
		return nil, fmt.Errorf("not found")
	}

	return ins[0], nil
}

func (c *Client) DiscoverInstancesWithName(name string, tags []string, passing int) ([]*ServiceInstance, error) {
	if passing == -1 {
		return c.DiscoverCatalogInstancesWithTags(name, tags)
	} else {
		return c.DiscoverHealthInstancesWithTags(name, tags, passing == 1)
	}
}

func (c *Client) DiscoverInstanceWithId(name string, id string, passing int) (*ServiceInstance, error) {
	if passing == -1 {
		return c.DiscoverCatalogInstanceWithId(name, id)
	} else {
		return c.DiscoverHealthInstanceWithId(name, id, passing == 1)
	}
}
