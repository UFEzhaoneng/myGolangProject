package consul

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"time"
)

//consul
// NewConsulRegister create a new consul register
func NewConsulRegister() *ConsulRegister {
	return &ConsulRegister{
		Address:                        "127.0.0.1:8500", //consul address
		Name:                           "unknown",
		Tag:                            []string{},
		Port:                           3000,
		DeregisterCriticalServiceAfter: time.Duration(1) * time.Minute,
		Interval:                       time.Duration(10) * time.Second,
	}
}

// ConsulRegister consul service register
type ConsulRegister struct {
	Address                        string
	Name                           string
	Tag                            []string
	Port                           int
	DeregisterCriticalServiceAfter time.Duration
	Interval                       time.Duration
}

func (r *ConsulRegister) register() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = r.Address
	client, err := api.NewClient(config)
	return client, err
}

// GRPCRegister register service
func (r *ConsulRegister) GRPCRegister() error {
	client, err := r.register()
	if err != nil {
		return err
	}
	agent := client.Agent()
	reg := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%v-%v-%v", r.Name, r.Address, r.Port), // 服务节点的名称
		Name:    r.Name,                                             // 服务名称
		Tags:    r.Tag,                                              // tag，可以为空
		Port:    r.Port,                                             // 服务端口
		Address: "127.0.0.1",                                        // 服务 IP
		Check: &api.AgentServiceCheck{ // 健康检查
			Interval:                       r.Interval.String(),                                // 健康检查间隔
			GRPC:                           fmt.Sprintf("%v:%v/%v", r.Address, r.Port, r.Name), // grpc 支持，执行健康检查的地址，service 会传到 Health.Check 函数中
			DeregisterCriticalServiceAfter: r.DeregisterCriticalServiceAfter.String(),          // 注销时间，相当于过期时间
		},
	}

	if err := agent.ServiceRegister(reg); err != nil {
		return err
	}

	return nil
}

// GRPCRegister register service
func (r *ConsulRegister) HttpRegister() error {
	client, err := r.register()
	if err != nil {
		return err
	}
	agent := client.Agent()
	reg := &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%v-%v-%v", r.Name, r.Address, r.Port), // 服务节点的名称
		Name:    r.Name,                                             // 服务名称
		Tags:    r.Tag,                                              // tag，可以为空
		Port:    r.Port,                                             // 服务端口
		Address: "127.0.0.1",                                        // 服务 IP
		Check: &api.AgentServiceCheck{ // 健康检查
			Interval:                       r.Interval.String(), // 健康检查间隔
			HTTP:                           fmt.Sprintf("http://%s:%d%s", r.Address, r.Port, "/check"),
			DeregisterCriticalServiceAfter: r.DeregisterCriticalServiceAfter.String(), // 注销时间，相当于过期时间
		},
	}

	if err := agent.ServiceRegister(reg); err != nil {
		return err
	}

	return nil
}
