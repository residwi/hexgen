-- Dev admin: admin@example.com / admin123456 (bcrypt cost 10).
-- These credentials are public. Never apply this seed to a production DATABASE_URL.
INSERT INTO users (email, password_hash, first_name, last_name, role, active) VALUES
('admin@example.com', '$2a$10$di3MUSPKPZiSdwcCVhRHtu09ZFeGfW29Ag6g6vlO65M7.rxNHOs5a', 'Admin', 'User', 'admin', true)
ON CONFLICT (email) DO NOTHING;
