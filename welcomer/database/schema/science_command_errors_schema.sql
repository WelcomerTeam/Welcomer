CREATE TABLE IF NOT EXISTS science_command_errors (
    command_uuid uuid NOT NULL UNIQUE PRIMARY KEY,
    trace text NOT NULL,
    data jsonb NULL,
    FOREIGN KEY (command_uuid) REFERENCES science_command_usages (command_uuid) ON DELETE CASCADE ON UPDATE CASCADE
);

