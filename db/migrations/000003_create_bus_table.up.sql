CREATE TABLE bus (
  id SERIAL NOT NULL UNIQUE,
	vehicle_no VARCHAR(32) NOT NULL,
  imei VARCHAR(32) NOT NULL UNIQUE,
  is_active BOOLEAN DEFAULT FALSE,
  color VARCHAR(16) DEFAULT 'biru',
  created_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP),
  updated_at BIGINT DEFAULT EXTRACT(EPOCH FROM CURRENT_TIMESTAMP),
  PRIMARY KEY (id, imei)
);

CREATE TRIGGER update_bus_updated_at BEFORE UPDATE ON bus FOR EACH ROW EXECUTE PROCEDURE update_modified_column();