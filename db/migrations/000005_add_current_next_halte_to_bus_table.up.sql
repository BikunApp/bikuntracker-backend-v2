-- Add current_halte and next_halte columns to bus table
ALTER TABLE bus ADD COLUMN current_halte VARCHAR(64);
ALTER TABLE bus ADD COLUMN next_halte VARCHAR(64);
