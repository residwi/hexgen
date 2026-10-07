-- Dev admin: admin@example.com / admin123456 (Argon2id, OWASP minimum parameters).
-- These credentials are public. Never apply this seed to a production DATABASE_URL.
INSERT INTO users (email, password_hash, first_name, last_name, role, active) VALUES
('admin@example.com', '$argon2id$v=19$m=19456,t=2,p=1$cZYZ0SFdcEg1lWw3StFAbg$BWfJtnOMWB3I6hpTwcEZzB/KDPM3EJDoT+tdCBRIkqc', 'Admin', 'User', 'admin', true)
ON CONFLICT (email) DO UPDATE SET password_hash = EXCLUDED.password_hash
WHERE users.password_hash NOT LIKE '$argon2id$%';
