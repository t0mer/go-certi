CREATE TABLE IF NOT EXISTS fqdn_channels (
    fqdn_id    TEXT NOT NULL REFERENCES fqdns(id) ON DELETE CASCADE,
    channel_id TEXT NOT NULL REFERENCES notification_channels(id) ON DELETE CASCADE,
    PRIMARY KEY (fqdn_id, channel_id)
);

CREATE INDEX IF NOT EXISTS idx_fqdn_channels_channel ON fqdn_channels(channel_id);
