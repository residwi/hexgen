-- Admin user (password: admin123)
INSERT INTO users (email, password_hash, first_name, last_name, role, active) VALUES
('admin@example.com', '$2a$10$77GRML5RGmBRYqO9g6JD5.2MjdptABU7DROSme4o6bUtwhfPKWb6.', 'Admin', 'User', 'admin', true)
ON CONFLICT (email) DO NOTHING;
