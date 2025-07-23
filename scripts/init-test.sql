-- Script de inicialización para base de datos de testing
CREATE DATABASE IF NOT EXISTS microservice_test;

-- Crear usuario para testing
CREATE USER IF NOT EXISTS 'test_user'@'%' IDENTIFIED BY 'test_password';
GRANT ALL PRIVILEGES ON microservice_test.* TO 'test_user'@'%';

-- Tablas para testing (misma estructura que desarrollo)
/*
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL PRIMARY KEY,
    user_id INTEGER,
    action VARCHAR(100) NOT NULL,
    details JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);
*/

-- Datos de prueba
/*
INSERT INTO users (email, name) VALUES 
    ('test@example.com', 'Test User'),
    ('admin@test.com', 'Test Admin');
*/