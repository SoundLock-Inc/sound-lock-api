-- Включение расширения для генерации UUID
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);


-- -- Таблица пользователей
-- CREATE TABLE users (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     email TEXT UNIQUE NOT NULL,
--     password TEXT NOT NULL,
--     created_at TIMESTAMP DEFAULT NOW(),
--     updated_at TIMESTAMP DEFAULT NOW()
-- );

-- -- Таблица аудиопаролей
-- CREATE TABLE audio_passwords (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
--     title TEXT, -- название пароля (по желанию пользователя)
--     password TEXT NOT NULL, -- зашифрованный сгенерированный пароль
--     length INT CHECK (length IN (8, 16, 32)) NOT NULL, -- длина генерируемого пароля
--     audio_path TEXT NOT NULL, -- путь до файла на сервере или в хранилище
--     created_at TIMESTAMP DEFAULT NOW(),
--     updated_at TIMESTAMP DEFAULT NOW()
-- );
