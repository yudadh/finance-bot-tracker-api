CREATE TABLE parser_attempts (
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT UNSIGNED NOT NULL,
    raw_text TEXT NOT NULL,
    parser_type ENUM ('rule_based', 'ai') NOT NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    confidence DECIMAL(5, 4) NULL,
    parsed_payload JSON NULL,
    error_message TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_parser_attempts_user FOREIGN KEY (user_id) REFERENCES users (id),
    INDEX idx_parser_attempts_user_created (user_id, created_at),
    INDEX idx_parser_attempts_parser_type (parser_type),
    INDEX idx_parser_attempts_success (success)
) ENGINE = InnoDB;