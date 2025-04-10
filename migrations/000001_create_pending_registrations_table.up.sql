CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS pending_registrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    otp VARCHAR(10) NOT NULL,
    otp_expiry TIMESTAMP NOT NULL,
    CONSTRAINT pending_registrations_email_unique UNIQUE (email),
    CONSTRAINT pending_registrations_phone_unique UNIQUE (phone)
);

CREATE INDEX idx_pending_registrations_email ON pending_registrations(email);
CREATE INDEX idx_pending_registrations_phone ON pending_registrations(phone);