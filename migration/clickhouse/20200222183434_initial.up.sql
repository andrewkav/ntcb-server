CREATE TABLE IF NOT EXISTS tracking.telemetry (
    device_id           FixedString(15),
    seq_no              UInt32,
    timestamp           DateTime,
    event_code          UInt16,
    status              UInt8,
    nav_valid           UInt8, -- boolean 0 or 1
    nav_satellite_count UInt8,
    nav_timestamp       DateTime,
    lon                 Float64,
    lat                 Float64,
    alt                 Float64,
    speed               Float32,
    direction           Float32,
    odometer            Float32,
    engine_rpm          UInt16,
    ignition_on         UInt8,
    fuel_level_liters   Float32,
    engine_temp         Int8,
    accel_position      UInt8,
    brake_position      UInt8,
    dist_until_service  Float32,
    details             String
)
    ENGINE ReplacingMergeTree() PARTITION BY toYYYYMM(timestamp) ORDER BY (device_id, nav_timestamp) SETTINGS index_granularity = 8192