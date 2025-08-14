package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"openpenpal-backend/internal/models"
)

// EnvelopeService 信封服务
type EnvelopeService struct {
	db        *gorm.DB
	creditSvc *CreditService
	userSvc   *UserService // 添加用户服务依赖
}

// NewEnvelopeService 创建信封服务实例
func NewEnvelopeService(db *gorm.DB) *EnvelopeService {
	return &EnvelopeService{
		db: db,
	}
}

// SetCreditService 设置积分服务
func (s *EnvelopeService) SetCreditService(creditSvc *CreditService) {
	s.creditSvc = creditSvc
}

// SetUserService 设置用户服务
func (s *EnvelopeService) SetUserService(userSvc *UserService) {
	s.userSvc = userSvc
}

// CreateDesign 创建信封设计
func (s *EnvelopeService) CreateDesign(userID string, req *models.CreateEnvelopeDesignRequest) (*models.EnvelopeDesign, error) {
	design := &models.EnvelopeDesign{
		ID:           uuid.New().String(),
		SchoolCode:   req.SchoolCode,
		Type:         req.Type,
		Theme:        req.Theme,
		ImageURL:     req.ImageURL,
		ThumbnailURL: req.ImageURL, // 暂时使用相同URL
		CreatorID:    userID,
		CreatorName:  "设计师", // TODO: 从用户信息获取
		Description:  req.Description,
		Status:       models.DesignStatusPending,
		Period:       req.Period,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.db.Create(design).Error; err != nil {
		return nil, fmt.Errorf("创建设计失败: %v", err)
	}

	return design, nil
}

// GetDesigns 获取信封设计列表 - 增强OP Code区域过滤
func (s *EnvelopeService) GetDesigns(filters map[string]interface{}) ([]models.EnvelopeDesign, error) {
	var designs []models.EnvelopeDesign
	query := s.db.Model(&models.EnvelopeDesign{})

	// FSD增强: 基于用户OP Code过滤可用信封
	if userID, ok := filters["user_id"].(string); ok && userID != "" && s.userSvc != nil {
		user, err := s.userSvc.GetUserByID(userID)
		if err == nil && user.OPCode != "" {
			// 构建查询条件：用户可以看到的信封
			// 1. 没有区域限制的信封
			// 2. 匹配用户OP Code前缀的信封
			// 3. 同城市的城市级信封
			userSchoolCode := ""
			if len(user.OPCode) >= 2 {
				userSchoolCode = user.OPCode[:2]
			}

			query = query.Where("supported_op_code_prefix = '' OR supported_op_code_prefix IS NULL OR "+
				"? LIKE CONCAT(supported_op_code_prefix, '%') OR "+
				"(type = 'city' AND school_code = ?)",
				user.OPCode, userSchoolCode)
		}
	}

	// 应用其他过滤条件
	if schoolCode, ok := filters["school_code"].(string); ok && schoolCode != "" {
		query = query.Where("school_code = ?", schoolCode)
	}
	if designType, ok := filters["type"].(string); ok && designType != "" {
		query = query.Where("type = ?", designType)
	}
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	// 默认只显示活跃和已批准的设计
	query = query.Where("is_active = ? AND status = ?", true, models.DesignStatusApproved)

	// 按投票数排序
	if err := query.Order("vote_count DESC").Find(&designs).Error; err != nil {
		return nil, fmt.Errorf("获取设计列表失败: %v", err)
	}

	return designs, nil
}

// VoteForDesign 为设计投票
func (s *EnvelopeService) VoteForDesign(userID, designID string) error {
	// 检查是否已投票
	var count int64
	if err := s.db.Model(&models.EnvelopeVote{}).
		Where("user_id = ? AND design_id = ?", userID, designID).
		Count(&count).Error; err != nil {
		return fmt.Errorf("检查投票状态失败: %v", err)
	}

	if count > 0 {
		return errors.New("您已经为此设计投过票了")
	}

	// 开始事务
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 创建投票记录
		vote := &models.EnvelopeVote{
			ID:        uuid.New().String(),
			DesignID:  designID,
			UserID:    userID,
			CreatedAt: time.Now(),
		}
		if err := tx.Create(vote).Error; err != nil {
			return fmt.Errorf("创建投票记录失败: %v", err)
		}

		// 更新设计的投票数
		if err := tx.Model(&models.EnvelopeDesign{}).
			Where("id = ?", designID).
			UpdateColumn("vote_count", gorm.Expr("vote_count + ?", 1)).Error; err != nil {
			return fmt.Errorf("更新投票数失败: %v", err)
		}

		return nil
	})
}

// GetUserEnvelopes 获取用户的信封
func (s *EnvelopeService) GetUserEnvelopes(userID string) ([]models.Envelope, error) {
	var envelopes []models.Envelope
	if err := s.db.Where("used_by = ?", userID).
		Preload("Design").
		Find(&envelopes).Error; err != nil {
		return nil, fmt.Errorf("获取用户信封失败: %v", err)
	}
	return envelopes, nil
}

// GetEnvelope 获取信封详情
func (s *EnvelopeService) GetEnvelope(envelopeID string) (*models.Envelope, error) {
	var envelope models.Envelope
	if err := s.db.Where("id = ?", envelopeID).
		Preload("Design").
		First(&envelope).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("信封不存在")
		}
		return nil, fmt.Errorf("获取信封失败: %v", err)
	}
	return &envelope, nil
}

// BindToLetter 绑定信封到信件
func (s *EnvelopeService) BindToLetter(envelopeID, letterID, userID string) error {
	// 检查信封是否属于用户且未使用
	var envelope models.Envelope
	if err := s.db.Where("id = ? AND used_by = ? AND status = ?",
		envelopeID, userID, models.EnvelopeStatusUnsent).
		First(&envelope).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("信封不存在或已使用")
		}
		return fmt.Errorf("查询信封失败: %v", err)
	}

	// 更新信封状态
	updates := map[string]interface{}{
		"status":    models.EnvelopeStatusUsed,
		"letter_id": letterID,
		"used_at":   time.Now(),
	}
	if err := s.db.Model(&envelope).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新信封状态失败: %v", err)
	}

	// 奖励绑定信封积分
	// TODO: 重新集成积分系统
	// if s.creditSvc != nil {
	// 	go func() {
	// 		if err := s.creditSvc.RewardEnvelopeBinding(userID, letterID); err != nil {
	// 			fmt.Printf("Failed to reward envelope binding: %v\n", err)
	// 		}
	// 	}()
	// }

	return nil
}

// VerifyEnvelope 验证信封条码
func (s *EnvelopeService) VerifyEnvelope(barcode string) (*models.Envelope, error) {
	var envelope models.Envelope
	if err := s.db.Where("barcode_id = ?", barcode).
		Preload("Design").
		First(&envelope).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("无效的信封条码")
		}
		return nil, fmt.Errorf("查询信封失败: %v", err)
	}
	return &envelope, nil
}

// GetEnvelopeByID 根据ID获取信封
func (s *EnvelopeService) GetEnvelopeByID(id string) (*models.Envelope, error) {
	return s.GetEnvelope(id)
}

// BindEnvelopeToLetter 绑定信封到信件
func (s *EnvelopeService) BindEnvelopeToLetter(envelopeID, letterID, userID string) error {
	return s.BindToLetter(envelopeID, letterID, userID)
}

// CreateEnvelopeOrder 创建信封订单 - 增强OP Code验证
func (s *EnvelopeService) CreateEnvelopeOrder(userID, designID string, quantity int) (*models.EnvelopeOrder, error) {
	// 验证设计是否存在且可用
	var design models.EnvelopeDesign
	if err := s.db.Where("id = ? AND status = ? AND is_active = ?",
		designID, models.DesignStatusApproved, true).First(&design).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("信封设计不存在或不可用")
		}
		return nil, fmt.Errorf("查询设计失败: %v", err)
	}

	// FSD增强: 验证用户OP Code与信封支持的区域
	if s.userSvc != nil && design.SupportedOPCodePrefix != "" {
		user, err := s.userSvc.GetUserByID(userID)
		if err != nil {
			return nil, fmt.Errorf("获取用户信息失败: %v", err)
		}

		// 检查用户是否有OP Code
		if user.OPCode == "" {
			return nil, errors.New("请先设置您的OP Code地址")
		}

		// 验证OP Code前缀匹配
		if len(user.OPCode) < len(design.SupportedOPCodePrefix) {
			return nil, errors.New("您的OP Code不完整")
		}

		userPrefix := user.OPCode[:len(design.SupportedOPCodePrefix)]
		if userPrefix != design.SupportedOPCodePrefix {
			// 城市级信封允许同城市的所有学校
			if design.Type == "city" && len(design.SchoolCode) >= 2 && len(user.OPCode) >= 2 {
				if user.OPCode[:2] != design.SchoolCode[:2] {
					return nil, fmt.Errorf("该信封仅限%s区域使用", design.SchoolCode[:2])
				}
			} else {
				return nil, fmt.Errorf("该信封仅限%s区域使用", design.SupportedOPCodePrefix)
			}
		}
	}

	// 计算价格 - 使用设计中的价格
	unitPrice := design.Price
	if unitPrice <= 0 {
		unitPrice = 2.0 // 默认价格
	}
	totalPrice := float64(quantity) * unitPrice

	// 创建订单
	order := &models.EnvelopeOrder{
		ID:             uuid.New().String(),
		UserID:         userID,
		DesignID:       designID,
		Quantity:       quantity,
		TotalPrice:     totalPrice,
		Status:         "pending",
		PaymentMethod:  "",
		DeliveryMethod: "pickup", // 默认自提
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := s.db.Create(order).Error; err != nil {
		return nil, fmt.Errorf("创建订单失败: %v", err)
	}

	// 奖励购买信封积分
	// TODO: 重新集成积分系统
	// if s.creditSvc != nil {
	// 	go func() {
	// 		if err := s.creditSvc.RewardEnvelopePurchase(userID, order.ID, quantity); err != nil {
	// 			fmt.Printf("Failed to reward envelope purchase: %v\n", err)
	// 		}
	// 	}()
	// }

	return order, nil
}

// GenerateEnvelopesForOrder 为订单生成信封
func (s *EnvelopeService) GenerateEnvelopesForOrder(orderID string) error {
	// 查询订单信息
	var order models.EnvelopeOrder
	if err := s.db.Where("id = ?", orderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("订单不存在")
		}
		return fmt.Errorf("查询订单失败: %v", err)
	}

	// 检查订单状态
	if order.Status != "paid" && order.Status != "pending" {
		return errors.New("订单状态不允许生成信封")
	}

	// 生成信封实例
	envelopes := make([]models.Envelope, 0, order.Quantity)
	for i := 0; i < order.Quantity; i++ {
		envelope := models.Envelope{
			ID:        uuid.New().String(),
			DesignID:  order.DesignID,
			UserID:    order.UserID,
			UsedBy:    order.UserID,
			BarcodeID: fmt.Sprintf("ENV-%s-%03d", orderID[0:8], i+1),
			Status:    models.EnvelopeStatusUnsent,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		envelopes = append(envelopes, envelope)
	}

	// 批量创建信封
	if err := s.db.CreateInBatches(envelopes, 100).Error; err != nil {
		return fmt.Errorf("批量创建信封失败: %v", err)
	}

	// 更新订单状态为已完成
	if err := s.db.Model(&order).Update("status", "completed").Error; err != nil {
		return fmt.Errorf("更新订单状态失败: %v", err)
	}

	fmt.Printf("Successfully generated %d envelopes for order: %s\n", order.Quantity, orderID)
	return nil
}

// GetUserEnvelopeOrders 获取用户的信封订单列表
func (s *EnvelopeService) GetUserEnvelopeOrders(userID string) ([]models.EnvelopeOrder, error) {
	var orders []models.EnvelopeOrder
	if err := s.db.Where("user_id = ?", userID).
		Preload("Design").
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("获取用户订单失败: %v", err)
	}
	return orders, nil
}

// GetEnvelopeOrder 获取订单详情
func (s *EnvelopeService) GetEnvelopeOrder(orderID string, userID string) (*models.EnvelopeOrder, error) {
	var order models.EnvelopeOrder
	if err := s.db.Where("id = ? AND user_id = ?", orderID, userID).
		Preload("Design").
		First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("订单不存在")
		}
		return nil, fmt.Errorf("查询订单失败: %v", err)
	}
	return &order, nil
}

// UpdateOrderPaymentStatus 更新订单支付状态
func (s *EnvelopeService) UpdateOrderPaymentStatus(orderID, paymentID, paymentMethod string) error {
	updates := map[string]interface{}{
		"status":         "paid",
		"payment_id":     paymentID,
		"payment_method": paymentMethod,
		"updated_at":     time.Now(),
	}

	if err := s.db.Model(&models.EnvelopeOrder{}).
		Where("id = ?", orderID).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("更新订单支付状态失败: %v", err)
	}
	return nil
}
