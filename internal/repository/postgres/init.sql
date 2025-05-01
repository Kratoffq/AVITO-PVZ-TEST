-- Создание таблицы пользователей
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Создание таблицы ПВЗ
CREATE TABLE IF NOT EXISTS pvzs (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    city VARCHAR(100) NOT NULL,
    CONSTRAINT city_check CHECK (city IN ('Москва', 'Санкт-Петербург', 'Казань'))
);

-- Создание таблицы приемок
CREATE TABLE IF NOT EXISTS receptions (
    id UUID PRIMARY KEY,
    date_time TIMESTAMP WITH TIME ZONE NOT NULL,
    pvz_id UUID NOT NULL REFERENCES pvzs(id),
    status VARCHAR(50) NOT NULL DEFAULT 'in_progress',
    CONSTRAINT status_check CHECK (status IN ('in_progress', 'close'))
);

-- Создание таблицы товаров
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    date_time TIMESTAMP WITH TIME ZONE NOT NULL,
    type VARCHAR(50) NOT NULL,
    reception_id UUID NOT NULL REFERENCES receptions(id),
    CONSTRAINT type_check CHECK (type IN ('электроника', 'одежда', 'обувь'))
);

-- Создание индексов
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_receptions_pvz_id ON receptions(pvz_id);
CREATE INDEX IF NOT EXISTS idx_receptions_status ON receptions(status);
CREATE INDEX IF NOT EXISTS idx_products_reception_id ON products(reception_id);
CREATE INDEX IF NOT EXISTS idx_receptions_date_time ON receptions(date_time);

-- Добавление комментариев к таблицам
COMMENT ON TABLE users IS 'Таблица пользователей системы';
COMMENT ON TABLE pvzs IS 'Таблица пунктов выдачи заказов';
COMMENT ON TABLE receptions IS 'Таблица приемок товаров';
COMMENT ON TABLE products IS 'Таблица товаров'; 