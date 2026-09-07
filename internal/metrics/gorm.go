package metrics

import "gorm.io/plugin/prometheus"

// GormPrometheusPlugin 返回 GORM 的 Prometheus 插件配置
func GormPrometheusPlugin(dbName string) *prometheus.Prometheus {
	return prometheus.New(prometheus.Config{
		DBName:          dbName, // 数据库名称，作为指标标签
		RefreshInterval: 15,     // 指标刷新频率（秒）
		// 如果希望 GORM 自动启动一个独立的 HTTP 服务暴露指标，可配置以下两项
		// 但这里我们统一使用上面的 Gin 独立端口，因此注释掉
		// StartServer:     true,
		// HTTPServerPort:  9090,
	})
}
