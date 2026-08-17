CREATE TABLE budgets (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    period_type ENUM ('weekly','monthly') NOT NULL DEFAULT 'monthly',
    period_start DATE NOT NULL,
    amount BIGINT NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'IDR',
    alert_80_sent_at DATETIME NULL,
    alert_100_sent_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_budgets_user FOREIGN KEY (user_id) REFERENCES users (id),
    UNIQUE KEY uq_budgets_user_period (user_id, period_type, period_start),
    INDEX idx_budgets_period (period_type, period_start)
) ENGINE = InnoDB;