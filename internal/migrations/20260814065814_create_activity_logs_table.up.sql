CREATE TABLE activity_logs (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NULL,
    admin_user_id BIGINT UNSIGNED NULL,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(100) NULL,
    entity_id BIGINT UNSIGNED NULL,
    metadata JSON NULL,
    ip_address VARCHAR(45) NULL,
    user_agent VARCHAR(500) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_activity_logs_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_activity_logs_admin_user FOREIGN KEY (admin_user_id) REFERENCES admin_users (id),
    INDEX idx_activity_logs_user_created (user_id, created_at),
    INDEX idx_activity_logs_admin_created (admin_user_id, created_at),
    INDEX idx_activity_logs_action (action)
) ENGINE = InnoDB;