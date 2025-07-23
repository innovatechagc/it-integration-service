-- Script de inicialización de base de datos para desarrollo
CREATE DATABASE IF NOT EXISTS microservice_dev;

-- Crear usuario específico para la aplicación
CREATE USER IF NOT EXISTS 'app_user'@'%' IDENTIFIED BY 'app_password';
GRANT ALL PRIVILEGES ON microservice_dev.* TO 'app_user'@'%';

-- Tablas de ejemplo (comentadas para testing)
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

-- Datos de ejemplo para desarrollo
/*
INSERT INTO users (email, name) VALUES 
    ('admin@example.com', 'Administrator'),
    ('user@example.com', 'Test User');
*/