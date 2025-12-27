-- +goose Up

-- customers table with rich attributes for future expansion
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Authentication (required)
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,

    -- Identity (required for KYC later)
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,

    -- Contact (optional initially)
    phone VARCHAR(20),
    phone_verified BOOLEAN DEFAULT FALSE,
    email_verified BOOLEAN DEFAULT FALSE,

    -- Address (optional, for KYC/compliance later)
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    postal_code VARCHAR(20),
    country CHAR(2),  -- ISO 3166-1 alpha-2 (NO, SE, US, etc.)

    -- Identity verification (for KYC later)
    date_of_birth DATE,
    national_id_number VARCHAR(50),  -- Encrypted in production!

    -- Status & security
    status VARCHAR(20) NOT NULL DEFAULT 'active',  -- active, suspended, closed
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMPTZ,
    last_login_at TIMESTAMPTZ,
    password_changed_at TIMESTAMPTZ,

    -- Preferences (for UI/notifications later)
    preferred_language CHAR(2) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',

    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for common lookups
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers (email);
CREATE INDEX IF NOT EXISTS idx_customers_status ON customers (status);
CREATE INDEX IF NOT EXISTS idx_customers_last_name ON customers (last_name);

-- Add customer foreign key to accounts table
ALTER TABLE accounts ADD COLUMN customer_id UUID REFERENCES customers(id);

-- Index for finding all accounts belonging to a customer
CREATE INDEX IF NOT EXISTS idx_accounts_customer_id ON accounts (customer_id);

-- +goose Down
DROP INDEX IF EXISTS idx_accounts_customer_id;
ALTER TABLE accounts DROP COLUMN IF EXISTS customer_id;

DROP INDEX IF EXISTS idx_customers_last_name;
DROP INDEX IF EXISTS idx_customers_status;
DROP INDEX IF EXISTS idx_customers_email;
DROP TABLE IF EXISTS customers;
