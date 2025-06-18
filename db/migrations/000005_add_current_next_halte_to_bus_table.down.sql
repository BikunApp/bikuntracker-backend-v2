-- Remove current_halte and next_halte columns from bus table
ALTER TABLE bus DROP COLUMN IF EXISTS current_halte;
ALTER TABLE bus DROP COLUMN IF EXISTS next_halte;
