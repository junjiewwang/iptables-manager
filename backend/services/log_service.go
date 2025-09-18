package services

import (
	"iptables-management-backend/config"
	"iptables-management-backend/models"
)

type LogService struct{}

// NewLogService 创建日志服务实例
func NewLogService() *LogService {
	return &LogService{}
}

// GetAllLogs 获取所有操作日志
func (s *LogService) GetAllLogs() ([]models.OperationLog, error) {
	var logs []models.OperationLog
	err := config.DB.Order("timestamp DESC").Find(&logs).Error
	return logs, err
}

// GetLogsByUser 根据用户获取操作日志
func (s *LogService) GetLogsByUser(username string) ([]models.OperationLog, error) {
	var logs []models.OperationLog
	err := config.DB.Where("username = ?", username).Order("timestamp DESC").Find(&logs).Error
	return logs, err
}

// CreateLog 创建操作日志
func (s *LogService) CreateLog(log *models.OperationLog) error {
	return config.DB.Create(log).Error
}

// LogOperation 记录操作日志的便捷方法
func (s *LogService) LogOperation(username, operation, details, ipAddress string) error {
	log := &models.OperationLog{
		Username:  username,
		Operation: operation,
		Details:   details,
		IPAddress: ipAddress,
	}
	return s.CreateLog(log)
}

// DeleteOldLogs 删除旧日志（保留最近30天）
func (s *LogService) DeleteOldLogs() error {
	return config.DB.Where("timestamp < DATE_SUB(NOW(), INTERVAL 30 DAY)").Delete(&models.OperationLog{}).Error
}