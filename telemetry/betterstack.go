package telemetry

// type BetterStackClient struct {
// 	endpoint string
// 	token    string
// 	client   *http.Client
// }
//
// type BetterStackMetric struct {
// 	DT    string                 `json:"dt"`
// 	Name  string                 `json:"name"`
// 	Gauge *BetterStackGaugeValue `json:"gauge,omitempty"`
// }
//
// type BetterStackGaugeValue struct {
// 	Value float64 `json:"value"`
// }
//
// func NewBetterStackClient(endpoint, token string) *BetterStackClient {
// 	return &BetterStackClient{
// 		endpoint: endpoint,
// 		token:    token,
// 		client: &http.Client{
// 			Timeout: 10 * time.Second,
// 		},
// 	}
// }
//
// func (c *BetterStackClient) SendGauge(ctx context.Context, name string, value float64) error {
// 	metric := BetterStackMetric{
// 		DT:   time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
// 		Name: name,
// 		Gauge: &BetterStackGaugeValue{
// 			Value: value,
// 		},
// 	}
//
// 	data, err := json.Marshal(metric)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal metric: %w", err)
// 	}
//
// 	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewBuffer(data))
// 	if err != nil {
// 		return fmt.Errorf("failed to create request: %w", err)
// 	}
//
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", "Bearer "+c.token)
//
// 	resp, err := c.client.Do(req)
// 	if err != nil {
// 		return fmt.Errorf("failed to send request: %w", err)
// 	}
// 	defer resp.Body.Close()
//
// 	if resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
// 	}
//
// 	return nil
// }
//
// // BetterStackMetrics wraps your existing metrics and sends to BetterStack
// type BetterStackMetrics struct {
// 	*ApplicationMetrics
// 	client *BetterStackClient
// }
//
// func NewBetterStackMetrics(appMetrics *ApplicationMetrics, endpoint, token string) *BetterStackMetrics {
// 	return &BetterStackMetrics{
// 		ApplicationMetrics: appMetrics,
// 		client:             NewBetterStackClient(endpoint, token),
// 	}
// }
//
// // SendHTTPRequestCount sends HTTP request count as gauge to BetterStack
// func (m *BetterStackMetrics) SendHTTPRequestCount(ctx context.Context, count float64) error {
// 	return m.client.SendGauge(ctx, "http_requests_total", count)
// }
//
// // SendAuthAttempts sends auth attempts as gauge to BetterStack
// func (m *BetterStackMetrics) SendAuthAttempts(ctx context.Context, count float64) error {
// 	return m.client.SendGauge(ctx, "auth_attempts_total", count)
// }
//
// // SendRegistrationCount sends registration count as gauge to BetterStack
// func (m *BetterStackMetrics) SendRegistrationCount(ctx context.Context, count float64) error {
// 	return m.client.SendGauge(ctx, "registrations_total", count)
// }
//
// // SendEmailsSent sends emails sent count as gauge to BetterStack
// func (m *BetterStackMetrics) SendEmailsSent(ctx context.Context, count float64) error {
// 	return m.client.SendGauge(ctx, "emails_sent_total", count)
// }
//
// // SendDBConnections sends active DB connections as gauge to BetterStack
// func (m *BetterStackMetrics) SendDBConnections(ctx context.Context, count float64) error {
// 	return m.client.SendGauge(ctx, "db_connections_active", count)
// }
//
// // SendBackgroundJobs sends background jobs count as gauge to BetterStack
// func (m *BetterStackMetrics) SendBackgroundJobs(ctx context.Context, count float64) error {
// 	return m.client.SendGauge(ctx, "background_jobs_total", count)
// }
//
// // RecordAndSendMetric records to OpenTelemetry and sends to BetterStack
// func (m *BetterStackMetrics) RecordAndSendMetric(ctx context.Context, name string, value int64, attrs ...attribute.KeyValue) error {
// 	// Send to BetterStack
// 	return m.client.SendGauge(ctx, name, float64(value))
// }
