-- +goose Up
-- +goose StatementBegin

-- Create a common function to update the updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop the function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- +goose StatementEnd