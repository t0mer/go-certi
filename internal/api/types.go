package api

// ---- FQDN ----

type CreateFQDNRequest struct {
	FQDN                 string   `json:"fqdn" binding:"required"`
	IncludeSubdomains    bool     `json:"include_subdomains"`
	Enabled              *bool    `json:"enabled"`
	NotificationsEnabled *bool    `json:"notifications_enabled"`
	ScheduleID           *string  `json:"schedule_id"`
	ChannelIDs           []string `json:"channel_ids"`
	NotificationEvents   []string `json:"notification_events"`
	ExpiryThresholdDays  *int     `json:"expiry_threshold_days"`
}

type UpdateFQDNRequest struct {
	FQDN                 string   `json:"fqdn" binding:"required"`
	IncludeSubdomains    bool     `json:"include_subdomains"`
	Enabled              bool     `json:"enabled"`
	NotificationsEnabled bool     `json:"notifications_enabled"`
	ScheduleID           *string  `json:"schedule_id"`
	ChannelIDs           []string `json:"channel_ids"`
	NotificationEvents   []string `json:"notification_events"`
	ExpiryThresholdDays  int      `json:"expiry_threshold_days"`
}

type FQDNResponse struct {
	ID                   string   `json:"id"`
	FQDN                 string   `json:"fqdn"`
	IncludeSubdomains    bool     `json:"include_subdomains"`
	Enabled              bool     `json:"enabled"`
	NotificationsEnabled bool     `json:"notifications_enabled"`
	ScheduleID           *string  `json:"schedule_id"`
	ChannelIDs           []string `json:"channel_ids"`
	NotificationEvents   []string `json:"notification_events"`
	ExpiryThresholdDays  int      `json:"expiry_threshold_days"`
	CreatedAt            string   `json:"created_at"`
	UpdatedAt            string   `json:"updated_at"`
}

// ---- Certificate ----

type CertificateResponse struct {
	ID           string   `json:"id"`
	FQDNID       string   `json:"fqdn_id"`
	Serial       string   `json:"serial"`
	IssuerCA     string   `json:"issuer_ca"`
	IssuerName   string   `json:"issuer_name"`
	SubjectCN    string   `json:"subject_cn"`
	SANs         []string `json:"sans"`
	NotBefore    string   `json:"not_before"`
	NotAfter     string   `json:"not_after"`
	DiscoveredAt string   `json:"discovered_at"`
	Source       string   `json:"source"`
	Revoked      bool     `json:"revoked"`
}

type CertificateListResponse struct {
	Items    []CertificateResponse `json:"items"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

// ---- Channel ----

type CreateChannelRequest struct {
	Name    string         `json:"name" binding:"required"`
	Type    string         `json:"type" binding:"required,oneof=shoutrrr greenapi waweb"`
	Config  map[string]any `json:"config" binding:"required"`
	Enabled *bool          `json:"enabled"`
}

type UpdateChannelRequest struct {
	Name    string         `json:"name" binding:"required"`
	Type    string         `json:"type" binding:"required,oneof=shoutrrr greenapi waweb"`
	Config  map[string]any `json:"config" binding:"required"`
	Enabled bool           `json:"enabled"`
}

// ---- Schedule ----

type CreateScheduleRequest struct {
	Name      string `json:"name" binding:"required"`
	CronExpr  string `json:"cron_expr" binding:"required"`
	IsDefault bool   `json:"is_default"`
	Enabled   *bool  `json:"enabled"`
}

type UpdateScheduleRequest struct {
	Name      string `json:"name" binding:"required"`
	CronExpr  string `json:"cron_expr" binding:"required"`
	IsDefault bool   `json:"is_default"`
	Enabled   bool   `json:"enabled"`
}

// ---- Settings ----

type UpdateSettingsRequest struct {
	AuthEnabled               bool    `json:"auth_enabled"`
	Username                  *string `json:"username"`
	Password                  *string `json:"password"`
	APITokenProtectionEnabled bool    `json:"api_token_protection_enabled"`
	Theme                     string  `json:"theme" binding:"omitempty,oneof=light dark system"`
	SslmateAPIKey             string  `json:"sslmate_api_key"`
	DefaultScheduleID         *string `json:"default_schedule_id"`
}

type SettingsResponse struct {
	AuthEnabled               bool    `json:"auth_enabled"`
	Username                  *string `json:"username"`
	APITokenProtectionEnabled bool    `json:"api_token_protection_enabled"`
	Theme                     string  `json:"theme"`
	SslmateAPIKey             string  `json:"sslmate_api_key"`
	DefaultScheduleID         *string `json:"default_schedule_id"`
}

// ---- Auth ----

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

// ---- Errors ----

type ErrorResponse struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}
