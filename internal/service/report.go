package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cjadmin/internal/models"

	"gorm.io/gorm"
)

// ReportService 自定义报表服务
type ReportService struct {
	db *gorm.DB
}

// NewReportService 创建报表服务实例
func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{db: db}
}

// 允许的数据源表
var allowedTables = map[string]string{
	"customers":         "客户表",
	"customer_accounts": "客户账户表",
	"orders":            "订单表",
	"invoices":          "发票表",
	"contracts":         "合同表",
	"activities":        "活动表",
	"banners":           "横幅表",
}

// 表字段元数据
var tableMetadata = map[string][]map[string]string{
	"customers": {
		{"field": "id", "name": "客户ID", "type": "int"},
		{"field": "customer_name", "name": "客户名称", "type": "string"},
		{"field": "customer_code", "name": "客户编码", "type": "string"},
		{"field": "contact_name", "name": "联系人", "type": "string"},
		{"field": "contact_phone", "name": "联系电话", "type": "string"},
		{"field": "email", "name": "邮箱", "type": "string"},
		{"field": "customer_level", "name": "客户等级", "type": "string"},
		{"field": "status", "name": "状态", "type": "string"},
		{"field": "created_at", "name": "创建时间", "type": "datetime"},
	},
	"customer_accounts": {
		{"field": "id", "name": "账户ID", "type": "int"},
		{"field": "customer_id", "name": "客户ID", "type": "int"},
		{"field": "fund_balance", "name": "资金余额", "type": "decimal"},
		{"field": "consumption_balance", "name": "消耗余额", "type": "decimal"},
		{"field": "gift_balance", "name": "赠送余额", "type": "decimal"},
		{"field": "credit_balance", "name": "授信余额", "type": "decimal"},
		{"field": "credit_limit", "name": "授信额度", "type": "decimal"},
	},
	"orders": {
		{"field": "id", "name": "订单ID", "type": "int"},
		{"field": "order_no", "name": "订单编号", "type": "string"},
		{"field": "customer_id", "name": "客户ID", "type": "int"},
		{"field": "order_type", "name": "订单类型", "type": "string"},
		{"field": "amount", "name": "金额", "type": "decimal"},
		{"field": "status", "name": "状态", "type": "string"},
		{"field": "rebate_rate", "name": "返点比例", "type": "decimal"},
		{"field": "rebate_amount", "name": "返点金额", "type": "decimal"},
		{"field": "created_at", "name": "创建时间", "type": "datetime"},
	},
}

// Create 创建报表定义（租户隔离）
func (s *ReportService) Create(report *models.CustomReport) error {
	if report.TenantID == 0 {
		return errors.New("租户ID不能为空")
	}
	if report.Name == "" {
		return errors.New("报表名称不能为空")
	}
	if report.BaseTable == "" {
		return errors.New("数据源不能为空")
	}

	// 验证数据源是否允许
	if _, ok := allowedTables[report.BaseTable]; !ok {
		return errors.New("不支持的数据源")
	}

	report.CreatedAt = time.Now()
	report.UpdatedAt = time.Now()

	return s.db.Create(report).Error
}

// Update 更新报表定义
func (s *ReportService) Update(tenantID, reportID int64, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()

	result := s.db.Model(&models.CustomReport{}).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", reportID, tenantID).
		Updates(updates)

	if result.RowsAffected == 0 {
		return errors.New("报表不存在")
	}
	return result.Error
}

// Delete 删除报表（软删除）
func (s *ReportService) Delete(tenantID, reportID int64) error {
	now := time.Now()
	result := s.db.Model(&models.CustomReport{}).
		Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", reportID, tenantID).
		Update("deleted_at", now)

	if result.RowsAffected == 0 {
		return errors.New("报表不存在")
	}
	return result.Error
}

// GetByID 获取报表详情
func (s *ReportService) GetByID(tenantID, reportID int64) (*models.CustomReport, error) {
	var report models.CustomReport
	err := s.db.Where("id = ? AND tenant_id = ? AND deleted_at IS NULL", reportID, tenantID).
		First(&report).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("报表不存在")
		}
		return nil, err
	}
	return &report, nil
}

// List 获取报表列表
func (s *ReportService) List(tenantID int64, page, pageSize int, filters map[string]interface{}) ([]models.CustomReport, int64, error) {
	var reports []models.CustomReport
	var total int64

	query := s.db.Model(&models.CustomReport{}).Where("tenant_id = ? AND deleted_at IS NULL", tenantID)

	// 应用筛选条件
	if category, ok := filters["category"].(string); ok && category != "" {
		query = query.Where("category = ?", category)
	}
	if name, ok := filters["name"].(string); ok && name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	if createdBy, ok := filters["created_by"].(int64); ok && createdBy > 0 {
		query = query.Where("created_by = ?", createdBy)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&reports).Error

	return reports, total, err
}

// Execute 执行报表查询
func (s *ReportService) Execute(tenantID, reportID, executedBy int64, parameters map[string]interface{}) ([]map[string]interface{}, int64, error) {
	startTime := time.Now()

	// 获取报表定义
	report, err := s.GetByID(tenantID, reportID)
	if err != nil {
		return nil, 0, err
	}

	// 构建SQL查询
	sql, args, err := s.buildQuery(tenantID, report, parameters)
	if err != nil {
		return nil, 0, err
	}

	// 执行查询
	var results []map[string]interface{}
	rows, err := s.db.Raw(sql, args...).Rows()
	if err != nil {
		return nil, 0, fmt.Errorf("查询执行失败: %v", err)
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return nil, 0, err
	}

	// 遍历结果
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, 0, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	rowCount := int64(len(results))
	executionTime := time.Since(startTime).Milliseconds()

	// 记录执行历史
	paramsJSON, _ := json.Marshal(parameters)
	execution := &models.ReportExecution{
		TenantID:        tenantID,
		ReportID:        reportID,
		ExecutedBy:      executedBy,
		ExecutedAt:      time.Now(),
		Parameters:      paramsJSON,
		RowCount:        int(rowCount),
		ExecutionTimeMs: int(executionTime),
		Status:          "completed",
	}
	s.db.Create(execution)

	return results, rowCount, nil
}

// buildQuery 构建SQL查询
func (s *ReportService) buildQuery(tenantID int64, report *models.CustomReport, parameters map[string]interface{}) (string, []interface{}, error) {
	var args []interface{}

	// 解析选择字段
	var selectFields []models.SelectField
	if err := json.Unmarshal(report.SelectFields, &selectFields); err != nil {
		return "", nil, fmt.Errorf("解析选择字段失败: %v", err)
	}

	// 构建SELECT部分
	var selectParts []string
	for _, field := range selectFields {
		var part string
		if field.Aggregate != "" {
			part = fmt.Sprintf("%s(%s)", field.Aggregate, field.Field)
		} else {
			part = field.Field
		}
		if field.Alias != "" {
			part += fmt.Sprintf(" AS \"%s\"", field.Alias)
		}
		selectParts = append(selectParts, part)
	}

	// 构建FROM部分
	sql := fmt.Sprintf("SELECT %s FROM %s", strings.Join(selectParts, ", "), report.BaseTable)

	// 构建JOIN部分
	if report.JoinConfig != nil {
		var joinConfigs []models.JoinConfig
		if err := json.Unmarshal(report.JoinConfig, &joinConfigs); err == nil {
			for _, join := range joinConfigs {
				sql += fmt.Sprintf(" %s JOIN %s ON %s", join.Type, join.Table, join.On)
			}
		}
	}

	// 构建WHERE部分 - 始终添加租户隔离
	sql += fmt.Sprintf(" WHERE %s.tenant_id = ?", report.BaseTable)
	args = append(args, tenantID)

	// 添加deleted_at过滤
	sql += fmt.Sprintf(" AND %s.deleted_at IS NULL", report.BaseTable)

	// 解析筛选条件
	if report.FilterConfig != nil {
		var filterConfigs []models.FilterConfig
		if err := json.Unmarshal(report.FilterConfig, &filterConfigs); err == nil {
			for _, filter := range filterConfigs {
				value := filter.Value
				if filter.IsParameter {
					// 从参数中获取值
					if paramValue, ok := parameters[filter.Field]; ok {
						value = paramValue
					} else {
						continue // 跳过未提供的参数
					}
				}

				if value == nil {
					continue
				}

				switch filter.Operator {
				case "=", "!=", ">", "<", ">=", "<=":
					sql += fmt.Sprintf(" AND %s %s ?", filter.Field, filter.Operator)
					args = append(args, value)
				case "LIKE":
					sql += fmt.Sprintf(" AND %s LIKE ?", filter.Field)
					args = append(args, "%"+fmt.Sprintf("%v", value)+"%")
				case "IN":
					sql += fmt.Sprintf(" AND %s IN (?)", filter.Field)
					args = append(args, value)
				case "BETWEEN":
					if values, ok := value.([]interface{}); ok && len(values) == 2 {
						sql += fmt.Sprintf(" AND %s BETWEEN ? AND ?", filter.Field)
						args = append(args, values[0], values[1])
					}
				}
			}
		}
	}

	// 构建GROUP BY部分
	if report.GroupByFields != nil {
		var groupByFields []string
		if err := json.Unmarshal(report.GroupByFields, &groupByFields); err == nil && len(groupByFields) > 0 {
			sql += " GROUP BY " + strings.Join(groupByFields, ", ")
		}
	}

	// 构建ORDER BY部分
	if report.OrderByConfig != nil {
		var orderByConfigs []models.OrderByConfig
		if err := json.Unmarshal(report.OrderByConfig, &orderByConfigs); err == nil && len(orderByConfigs) > 0 {
			var orderParts []string
			for _, order := range orderByConfigs {
				orderParts = append(orderParts, fmt.Sprintf("%s %s", order.Field, order.Direction))
			}
			sql += " ORDER BY " + strings.Join(orderParts, ", ")
		}
	}

	// 添加LIMIT防止数据量过大
	sql += " LIMIT 10000"

	return sql, args, nil
}

// GetDataSources 获取可用数据源列表
func (s *ReportService) GetDataSources() []map[string]string {
	var sources []map[string]string
	for table, name := range allowedTables {
		sources = append(sources, map[string]string{
			"table": table,
			"name":  name,
		})
	}
	return sources
}

// GetTableFields 获取表字段元数据
func (s *ReportService) GetTableFields(tableName string) ([]map[string]string, error) {
	fields, ok := tableMetadata[tableName]
	if !ok {
		return nil, errors.New("不支持的数据表")
	}
	return fields, nil
}

// GetExecutionHistory 获取执行历史
func (s *ReportService) GetExecutionHistory(tenantID, reportID int64, page, pageSize int) ([]models.ReportExecution, int64, error) {
	var executions []models.ReportExecution
	var total int64

	query := s.db.Model(&models.ReportExecution{}).
		Where("tenant_id = ? AND report_id = ?", tenantID, reportID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("executed_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&executions).Error

	return executions, total, err
}

// GetStatistics 获取报表统计
func (s *ReportService) GetStatistics(tenantID int64) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 总报表数
	var totalCount int64
	s.db.Model(&models.CustomReport{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Count(&totalCount)
	stats["total_count"] = totalCount

	// 今日执行次数
	today := time.Now().Truncate(24 * time.Hour)
	var todayExecutions int64
	s.db.Model(&models.ReportExecution{}).
		Where("tenant_id = ? AND executed_at >= ?", tenantID, today).
		Count(&todayExecutions)
	stats["today_executions"] = todayExecutions

	// 总执行次数
	var totalExecutions int64
	s.db.Model(&models.ReportExecution{}).
		Where("tenant_id = ?", tenantID).
		Count(&totalExecutions)
	stats["total_executions"] = totalExecutions

	// 按类别统计
	var categoryStats []struct {
		Category string
		Count    int64
	}
	s.db.Model(&models.CustomReport{}).
		Select("category, COUNT(*) as count").
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Group("category").
		Scan(&categoryStats)
	stats["category_stats"] = categoryStats

	return stats, nil
}
