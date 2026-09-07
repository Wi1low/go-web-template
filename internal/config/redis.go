package config

import "time"

type Redis struct {
	Addr          string        `mapstructure:"addr"`           // 服务器地址:端口
	Password      string        `mapstructure:"password"`       // 密码
	DB            int           `mapstructure:"db"`             // 数据库编号
	PoolSize      int           `mapstructure:"pool-size"`      // 连接池最大连接数 [citation:7]
	MinIdleConns  int           `mapstructure:"min-idle-conns"` // 最小空闲连接数 [citation:7]
	MaxIdleConns  int           `mapstructure:"max-idle-conns"` // 最大空闲连接数（部分客户端用）[citation:7]
	IdleTimeout   time.Duration `mapstructure:"idle-timeout"`   // 空闲连接超时时间 [citation:9]
	DialTimeout   time.Duration `mapstructure:"dial-timeout"`   // 连接超时 [citation:9]
	ReadTimeout   time.Duration `mapstructure:"read-timeout"`   // 读超时 [citation:9]
	WriteTimeout  time.Duration `mapstructure:"write-timeout"`  // 写超时 [citation:9]
	MaxRetries    int           `mapstructure:"max-retries"`    // 最大重试次数 [citation:7]
	MasterName    string        `mapstructure:"master-name"`    // 哨兵模式下主节点名称 [citation:9]
	SentinelAddrs []string      `mapstructure:"sentinel-addrs"` // 哨兵地址列表 [citation:9]
	ClusterMode   bool          `mapstructure:"cluster-mode"`   // 是否集群模式 [citation:9]
}
