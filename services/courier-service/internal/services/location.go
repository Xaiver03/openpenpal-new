package services

import (
	"courier-service/internal/models"
	"fmt"
	"math"
)

// LocationService 地理位置服务
type LocationService struct{}

// NewLocationService 创建地理位置服务实例
func NewLocationService() *LocationService {
	return &LocationService{}
}

// CalculateDistance 使用Haversine公式计算两点之间的距离（单位：公里）
func (s *LocationService) CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // 地球半径 (km)

	// 转换为弧度
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// FindNearbyTasks 查找附近的任务
func (s *LocationService) FindNearbyTasks(courierLat, courierLng float64, radiusKm float64, tasks []models.Task) []models.Task {
	var nearbyTasks []models.Task

	for _, task := range tasks {
		if task.PickupLat != 0 && task.PickupLng != 0 {
			distance := s.CalculateDistance(courierLat, courierLng, task.PickupLat, task.PickupLng)
			if distance <= radiusKm {
				// 添加距离信息到任务中
				task.EstimatedDistance = s.FormatDistance(distance)
				nearbyTasks = append(nearbyTasks, task)
			}
		}
	}

	return nearbyTasks
}

// EstimateDeliveryTime 估算投递时间
func (s *LocationService) EstimateDeliveryTime(pickupLat, pickupLng, deliveryLat, deliveryLng float64) string {
	distance := s.CalculateDistance(pickupLat, pickupLng, deliveryLat, deliveryLng)

	// 简单的时间估算：假设平均速度15km/h（考虑校园内步行+等待时间）
	timeHours := distance / 15.0

	if timeHours < 0.5 {
		return "30分钟内"
	} else if timeHours < 1.0 {
		return "1小时内"
	} else if timeHours < 2.0 {
		return "2小时内"
	} else {
		return "3小时以上"
	}
}

// CalculateReward 根据距离计算奖励
func (s *LocationService) CalculateReward(pickupLat, pickupLng, deliveryLat, deliveryLng float64, priority string) float64 {
	distance := s.CalculateDistance(pickupLat, pickupLng, deliveryLat, deliveryLng)

	// 基础奖励
	baseReward := 5.0

	// 距离奖励：每公里增加1元
	distanceReward := distance * 1.0

	// 优先级奖励
	priorityReward := 0.0
	switch priority {
	case models.TaskPriorityUrgent:
		priorityReward = 3.0
	case models.TaskPriorityExpress:
		priorityReward = 5.0
	}

	total := baseReward + distanceReward + priorityReward

	// 最低奖励3元，最高奖励50元
	if total < 3.0 {
		total = 3.0
	} else if total > 50.0 {
		total = 50.0
	}

	return math.Round(total*100) / 100 // 保留两位小数
}

// FormatDistance 格式化距离显示
func (s *LocationService) FormatDistance(distanceKm float64) string {
	if distanceKm < 1.0 {
		meters := int(distanceKm * 1000)
		return fmt.Sprintf("%dm", meters)
	} else {
		return fmt.Sprintf("%.1fkm", distanceKm)
	}
}

// ParseLocation 解析位置字符串（这里可以集成地理编码服务）
func (s *LocationService) ParseLocation(location string) (float64, float64, error) {
	// 这里可以集成百度地图、高德地图等地理编码API
	// 暂时返回默认值

	// 北京大学的大致坐标
	if location == "北京大学" || location == "北大" {
		return 39.9912, 116.3064, nil
	}

	// 清华大学的大致坐标
	if location == "清华大学" || location == "清华" {
		return 40.0038, 116.3265, nil
	}

	// 默认返回北京市中心
	return 39.9042, 116.4074, nil
}

// GetZoneFromCoordinate 根据坐标获取所属区域
func (s *LocationService) GetZoneFromCoordinate(lat, lng float64) string {
	// 这里可以实现更复杂的区域判断逻辑
	// 暂时简单处理

	// 北京大学范围
	if lat >= 39.985 && lat <= 39.997 && lng >= 116.300 && lng <= 116.315 {
		return "北京大学"
	}

	// 清华大学范围
	if lat >= 39.998 && lat <= 40.010 && lng >= 116.320 && lng <= 116.335 {
		return "清华大学"
	}

	return "其他区域"
}
