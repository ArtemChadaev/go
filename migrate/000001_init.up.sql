CREATE TABLE users
(
    id            SERIAL       NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    CONSTRAINT users_pk PRIMARY KEY (id)
);
-- Токены для разных устройств
CREATE TABLE user_refresh_tokens
(
    id          SERIAL PRIMARY KEY,
    user_id     INT          NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    token       VARCHAR(255) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ  NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    name_device VARCHAR(255),
    device_info VARCHAR(255) -- Полезно для отладки
);
-- Функция для обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';
-- Тригер функции выше
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
EXECUTE PROCEDURE update_updated_at_column();

-- Настройки пользователя
CREATE TABLE user_settings
(
    user_id                   INT         NOT NULL UNIQUE,
    name                      VARCHAR(255) DEFAULT 'Alex',
    icon                      VARCHAR(255),
    coin                      INT DEFAULT 0,
    date_of_registration      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    paid_subscription         boolean DEFAULT FALSE,
--     Дата ОКОНЧАНИЯ её
    date_of_paid_subscription TIMESTAMPTZ,
    CONSTRAINT user_settings_pk PRIMARY KEY (user_id),
    CONSTRAINT fk_user_settings_user_id
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE TABLE clan
(
    id          SERIAL       NOT NULL UNIQUE,
    name        VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    other       JSONB
);
-- Роли для кланов
CREATE TABLE roles
(
    -- Чем ниже id тем выше ранг 1-самый лучший
    id   SMALLINT    NOT NULL UNIQUE,
    name VARCHAR(20) NOT NULL
);
-- Кланы до n людей в каждом
CREATE TABLE clan_members
(
    -- ID клана, к которому принадлежит участник
    clan_id INT      NOT NULL,
    -- ID самого пользователя-участника
    user_id INT      NOT NULL,
    -- Роль пользователя в этом конкретном клане (например, "лидер", "офицер", "рядовой")
    role_id SMALLINT NOT NULL,

    -- Создаем внешний ключ для clan_id, который ссылается на таблицу clan
    CONSTRAINT fk_clan
        FOREIGN KEY (clan_id)
            REFERENCES clan (id)
            ON DELETE CASCADE,

    -- Создаем внешний ключ для user_id, который ссылается на таблицу users
    CONSTRAINT fk_user
        FOREIGN KEY (user_id)
            REFERENCES users (id)
            ON DELETE CASCADE,

    -- Ссылка на наш новый словарь ролей
    CONSTRAINT fk_role
        FOREIGN KEY (role_id)
            REFERENCES roles (id),

    -- Делаем так, чтобы пара (clan_id, user_id) была уникальной.
    PRIMARY KEY (clan_id, user_id)
);
-- Вносятся роли
INSERT INTO roles (id, name)
VALUES (1, 'First'),
       (2, 'Second'),
       (3, 'Third'),
       (4, 'Fourth'),
       (5, 'Fifth');
-- Таблица для хранения названий ролей, заданных кланами
CREATE TABLE clan_role_names
(
    -- ID клана, который задает название
    clan_id     INT         NOT NULL,
    -- ID системной роли (ссылается на наш словарь)
    role_id     SMALLINT    NOT NULL,
    -- Название, которое придумал клан
    custom_name VARCHAR(20) NOT NULL,

    -- Внешний ключ, ссылающийся на клан
    CONSTRAINT fk_clan
        FOREIGN KEY (clan_id)
            REFERENCES clan (id)
            ON DELETE CASCADE,

    -- Внешний ключ, ссылающийся на словарь ролей
    CONSTRAINT fk_role
        FOREIGN KEY (role_id)
            REFERENCES roles (id)
            ON DELETE CASCADE,

    -- Уникальная пара, чтобы клан не мог задать два названия для одной и той же роли
    PRIMARY KEY (clan_id, role_id)
);
CREATE TABLE cards
(
    id          SERIAL       NOT NULL UNIQUE,
    user_id     INT          NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    other       jsonb,
    CONSTRAINT user_cards_pk PRIMARY KEY (id),
    CONSTRAINT fk_user_cards_user_id
        FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);
CREATE TABLE items
(
    id          SERIAL       NOT NULL UNIQUE,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    HaveCard    boolean,
    other       jsonb
);