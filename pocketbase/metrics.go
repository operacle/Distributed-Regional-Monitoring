
package pocketbase

func (c *PocketBaseClient) SaveMetrics(metrics MetricsRecord) error {
	return c.createRecord("services_metrics", metrics)
}

func (c *PocketBaseClient) SavePingData(pingData PingDataRecord) error {
	return c.createRecord("ping_data", pingData)
}

func (c *PocketBaseClient) SaveUptimeData(uptimeData UptimeDataRecord) error {
	return c.createRecord("uptime_data", uptimeData)
}

func (c *PocketBaseClient) SaveDNSData(dnsData DNSDataRecord) error {
	return c.createRecord("dns_data", dnsData)
}

func (c *PocketBaseClient) SaveTCPData(tcpData TCPDataRecord) error {
	return c.createRecord("tcp_data", tcpData)
}
