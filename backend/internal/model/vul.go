package model

import (
	"time"

	"gorm.io/gorm"
)

// VulEnv 漏洞环境模型
type VulEnv struct {
	gorm.Model
	EnvName     string    `gorm:"type:varchar(100);not null;comment:环境名称;uniqueIndex"`
	EnvDesc     string    `gorm:"type:text;comment:环境描述"`
	EnvType     string    `gorm:"type:varchar(50);not null;comment:环境类型(单镜像/复合环境)"`
	BaseImage   string    `gorm:"type:varchar(255);comment:基础镜像名称"`
	BaseCompose string    `gorm:"type:varchar(255);comment:docker-compose文件路径"`
	Rank        float64   `gorm:"type:decimal(3,1);default:3.5;comment:环境评分"`
	From        string    `gorm:"type:varchar(100);comment:来源"`
	CreateTime  time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdateTime  time.Time `gorm:"type:TIMESTAMP;default:CURRENT_TIMESTAMP;comment:更新时间"`
	Degree      string    `gorm:"type:json;comment:环境信息"`
	IsOpen      int       `gorm:"type:int;default:1000;comment:开放级别"`
	Cost        float64   `gorm:"type:decimal(10,2);default:0.00;comment:开启环境的成本"`
}

// VulInstance 用户开启的漏洞环境记录
type VulInstance struct {
	gorm.Model
	VulEnv      VulEnv    `gorm:"foreignKey:VulEnvID"`
	UserID      uint      `gorm:"not null;index;comment:用户ID"`
	VulEnvID    uint      `gorm:"not null;index;comment:漏洞环境ID"`
	StartTime   time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;comment:开启时间"`
	EndTime     time.Time `gorm:"type:datetime;comment:结束时间"`
	Status      int       `gorm:"type:tinyint;default:0;comment:状态(0/未创建/1运行中/2已停止/3已完成)"`
	StackName   string    `gorm:"type:varchar(100);comment:Docker Stack名称"`
	ContainerID string    `gorm:"type:varchar(64);comment:容器ID"`
	Ports       string    `gorm:"type:json;comment:端口映射"`
	ExpireTime  time.Time `gorm:"type:datetime;comment:过期时间"`
}

// VulEnv CRUD 操作
func CreateVulEnv(vul *VulEnv) error {
	return DB.Create(vul).Error
}

func GetVulEnvByID(id uint) (*VulEnv, error) {
	var vul VulEnv
	err := DB.First(&vul, id).Error
	if err != nil {
		return nil, err
	}
	return &vul, nil
}

func GetVulEnvByName(name string) (*VulEnv, error) {
	var vul VulEnv
	err := DB.Where("env_name = ?", name).First(&vul).Error
	if err != nil {
		return nil, err
	}
	return &vul, nil
}

// GetAllVulEnvs 分页查询漏洞环境
func GetAllVulEnvs(page, pageSize int) ([]VulEnv, error) {
	var vulEnvs []VulEnv
	offset := (page - 1) * pageSize
	err := DB.Offset(offset).Limit(pageSize).Find(&vulEnvs).Error
	return vulEnvs, err
}

// 不分页查询
func GetAllVulEnvsNoPage() ([]VulEnv, error) {
	var vulEnvs []VulEnv
	err := DB.Find(&vulEnvs).Error
	return vulEnvs, err
}

// GetVulEnvsByOpenLevel 根据开放级别查询漏洞环境
func GetVulEnvsByOpenLevel(roleLevel int) ([]VulEnv, error) {
	var vulEnvs []VulEnv
	err := DB.Where("is_open > ?", roleLevel).Find(&vulEnvs).Error
	if err != nil {
		return nil, err
	}
	return vulEnvs, nil
}

func UpdateVulEnv(vul *VulEnv) error {
	return DB.Save(vul).Error
}

func DeleteVulEnv(id uint) error {
	return DB.Unscoped().Delete(&VulEnv{}, id).Error
}

// VulInstance CRUD 操作
func CreateVulInstance(userVul *VulInstance) error {
	return DB.Create(userVul).Error
}

func GetVulInstanceByID(id uint) (*VulInstance, error) {
	var userVul VulInstance
	err := DB.First(&userVul, id).Error
	if err != nil {
		return nil, err
	}
	return &userVul, nil
}

func GetVulInstanceByUserID(userID uint) ([]VulInstance, error) {
	var userVuls []VulInstance
	err := DB.Where("user_id = ?", userID).Find(&userVuls).Error
	return userVuls, err
}

func GetVulInstanceByVulEnvID(vulEnvID uint) ([]VulInstance, error) {
	var userVuls []VulInstance
	err := DB.Where("vul_env_id =?", vulEnvID).Find(&userVuls).Error
	return userVuls, err
}

func UpdateVulInstance(userVul *VulInstance) error {
	return DB.Save(userVul).Error
}

func DeleteVulInstance(id uint) error {
	return DB.Delete(&VulInstance{}, id).Error
}

func GetVulInstanceBy2ID(userID, vulEnvID uint) (*VulInstance, error) {
	var userVul VulInstance
	err := DB.Where("user_id = ? AND vul_env_id = ?", userID, vulEnvID).First(&userVul).Error
	if err != nil {
		return nil, err
	}
	return &userVul, nil
}

// GetAllVulInstances 获取所有漏洞环境实例记录(不分页)
func GetAllVulInstances() ([]VulInstance, error) {
	var instances []VulInstance
	err := DB.Find(&instances).Error
	return instances, err
}

// DeleteVulInstanceByVulEnvID 根据漏洞环境ID删除漏洞实例(硬删除)
func DeleteVulInstanceByVulEnvID(vulEnvID uint) error {
	result := DB.Unscoped().Where("vul_env_id = ?", vulEnvID).Delete(&VulInstance{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// DeleteVulInstanceBy2ID 根据用户ID和漏洞环境ID删除漏洞实例
func DeleteVulInstanceBy2ID(userID, vulEnvID uint) error {
	result := DB.Where("user_id = ? AND vul_env_id = ?", userID, vulEnvID).Delete(&VulInstance{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ExtendExpireTime 延长实例过期时间半小时
func ExtendExpireTime(id uint) error {
	// 获取实例
	instance, err := GetVulInstanceByID(id)
	if err != nil {
		return err
	}

	// 延长半小时
	instance.ExpireTime = instance.ExpireTime.Add(30 * time.Minute)

	// 更新数据库
	return UpdateVulInstance(instance)
}
