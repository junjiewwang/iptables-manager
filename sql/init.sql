-- IPTables管理系统数据库初始化脚本

-- 创建用户表
CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(50) NOT NULL,
  `password` varchar(255) NOT NULL,
  `role` varchar(20) DEFAULT 'user',
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_users_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 创建IPTables规则表
CREATE TABLE IF NOT EXISTS `iptables_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `chain_name` varchar(50) NOT NULL,
  `rule_number` bigint DEFAULT NULL,
  `target` varchar(20) NOT NULL,
  `protocol` varchar(10) DEFAULT NULL,
  `source_ip` varchar(45) DEFAULT NULL,
  `destination_ip` varchar(45) DEFAULT NULL,
  `source_port` varchar(20) DEFAULT NULL,
  `destination_port` varchar(20) DEFAULT NULL,
  `interface_in` varchar(20) DEFAULT NULL,
  `interface_out` varchar(20) DEFAULT NULL,
  `rule_text` text,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_iptables_rules_rule_number` (`rule_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 创建操作日志表
CREATE TABLE IF NOT EXISTS `operation_logs` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `username` varchar(50) NOT NULL,
  `operation` varchar(100) NOT NULL,
  `details` text,
  `timestamp` datetime(3) DEFAULT NULL,
  `ip_address` varchar(45) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_operation_logs_username` (`username`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- 插入默认用户数据
INSERT IGNORE INTO `users` (`username`, `password`, `role`, `created_at`, `updated_at`) VALUES
('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', NOW(), NOW()),
('user1', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'user', NOW(), NOW());

-- 插入示例规则数据
INSERT IGNORE INTO `iptables_rules` (`chain_name`, `target`, `protocol`, `source_ip`, `destination_ip`, `destination_port`, `rule_text`, `created_at`, `updated_at`) VALUES
('INPUT', 'ACCEPT', 'tcp', '0.0.0.0/0', '0.0.0.0/0', '22', 'iptables -A INPUT -p tcp --dport 22 -j ACCEPT', NOW(), NOW()),
('INPUT', 'ACCEPT', 'tcp', '0.0.0.0/0', '0.0.0.0/0', '80', 'iptables -A INPUT -p tcp --dport 80 -j ACCEPT', NOW(), NOW()),
('INPUT', 'ACCEPT', 'tcp', '0.0.0.0/0', '0.0.0.0/0', '443', 'iptables -A INPUT -p tcp --dport 443 -j ACCEPT', NOW(), NOW()),
('INPUT', 'DROP', '', '0.0.0.0/0', '0.0.0.0/0', '', 'iptables -A INPUT -j DROP', NOW(), NOW());

-- 插入示例操作日志
INSERT IGNORE INTO `operation_logs` (`username`, `operation`, `details`, `timestamp`, `ip_address`) VALUES
('admin', '系统初始化', '创建默认规则和用户', NOW(), '127.0.0.1'),
('admin', '创建规则', '允许SSH访问', NOW(), '127.0.0.1'),
('admin', '创建规则', '允许HTTP访问', NOW(), '127.0.0.1'),
('admin', '创建规则', '允许HTTPS访问', NOW(), '127.0.0.1');