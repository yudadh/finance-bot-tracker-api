CREATE table users (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    telegram_id BIGINT NOT NULL UNIQUE,
    telegram_username VARCHAR(255) NULL,
    first_name VARCHAR(255) NULL,
    last_name VARCHAR(255) NULL,
    language_code VARCHAR(10) NOT NULL DEFAULT 'id',
    timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Makassar',
    status ENUM ('active', 'inactive', 'blocked') NOT NULL DEFAULT 'active',
    last_seen_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) engine = InnoDb;