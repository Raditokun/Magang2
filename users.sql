
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
	role VARCHAR(50),
	nip INT,
	email VARCHAR(100),
	password VARCHAR(100),
	status VARCHAR(100),
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    updated_at TIMESTAMP,
    updated_by VARCHAR(100),
);



