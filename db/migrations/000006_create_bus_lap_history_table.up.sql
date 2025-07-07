-- Create table to log history of bus laps
CREATE TABLE bus_lap_history (
    id SERIAL PRIMARY KEY,
    bus_id INTEGER NOT NULL REFERENCES bus(id) ON DELETE CASCADE,
    imei VARCHAR(32) NOT NULL,
    lap_number INTEGER NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    route_color VARCHAR(16),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_bus_lap_history_bus_id ON bus_lap_history(bus_id);
CREATE INDEX idx_bus_lap_history_imei ON bus_lap_history(imei);
