-- Add operation and metadata columns to logs table for enhanced filtering and context
ALTER TABLE logs ADD COLUMN operation TEXT;
ALTER TABLE logs ADD COLUMN metadata TEXT;

-- Create index on operation for efficient filtering by operation type
CREATE INDEX IF NOT EXISTS idx_logs_operation ON logs(operation);
