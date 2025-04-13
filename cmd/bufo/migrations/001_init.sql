CREATE TABLE IF NOT EXISTS internal (
				id SERIAL PRIMARY KEY, 
				name TEXT NOT NULL, 
				value TEXT NOT NULL
);

INSERT INTO internal (name, value) VALUES ('version', '0.1');
