CREATE TABLE categories (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    type ENUM ('income', 'expense') NOT NULL,
    keywords JSON NULL,
    is_default BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME NULL,
    UNIQUE KEY uq_categories_name_type_deleted_at (name, type, deleted_at),
    INDEX idx_categories_type (type),
    INDEX idx_categories_deleted_at (deleted_at)
) ENGINE = InnoDB;