CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
   IF row(NEW.*) IS DISTINCT FROM row(OLD.*) THEN
      NEW.updated_at = EXTRACT(EPOCH FROM CURRENT_TIMESTAMP);
      RETURN NEW;
   ELSE
      RETURN OLD;
   END IF;
END;
$$ language 'plpgsql';
