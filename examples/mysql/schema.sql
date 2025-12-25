-- MySQL Schema for toutago-datamapper Example

-- Create database (if running manually)
CREATE DATABASE IF NOT EXISTS testdb;
USE testdb;

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    version INT DEFAULT 1,
    INDEX idx_email (email),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    stock INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_price (price)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Optional: Create sample data for testing
-- INSERT INTO users (name, email) VALUES 
--     ('John Doe', 'john@example.com'),
--     ('Jane Smith', 'jane@example.com');
-- 
-- INSERT INTO products (name, description, price, stock) VALUES 
--     ('Widget', 'A useful widget', 19.99, 100),
--     ('Gadget', 'An amazing gadget', 49.99, 50);
