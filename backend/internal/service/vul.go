package service

import (
	"AscensionPath/config"
	"AscensionPath/internal/middleware"
	"AscensionPath/internal/model"
	"AscensionPath/internal/utils"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

func init() {
	StartMonitorExpiredInstances()
}

type VulService struct{}

type VulImagesList []VulImage

type VulDegree struct {
	HoleType    []string `json:"HoleType,omitempty"`
	DevClassify []string `json:"devClassify,omitempty"`
	DevDatabase []string `json:"devDatabase,omitempty"`
	DevLanguage []string `json:"devLanguage,omitempty"`
}

type VulImage struct {
	ImageName    string    `json:"image_name"`
	ImageVulName string    `json:"image_vul_name"`
	ImageDesc    string    `json:"image_desc"`
	Rank         float64   `json:"rank"`
	Degree       VulDegree `json:"degree"`
	From         string    `json:"from"`
}

// ReadJsonFile 读取 JSON 文件并返回 VulImagesList
func (v *VulService) ReadJsonFile(filePath string) (VulImagesList, error) {
	var vulImagesList VulImagesList
	err := utils.ReadJson(filePath, &vulImagesList)
	if err != nil {
		return nil, err
	}
	return vulImagesList, nil
}

// 从本地路径获取漏洞镜像信息
func (v *VulService) GetVulImageFromLocal(path string) (VulImagesList, error) {
	// 检查目录是否存在，不存在则创建
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, fmt.Errorf("创建目录失败: %v", err)
		}
	}
	// 读取目录中的所有文件
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %v", err)
	}

	var result VulImagesList

	// 遍历目录中的文件
	for _, file := range files {
		// 只处理.json文件
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// 构建完整文件路径
		filePath := filepath.Join(path, file.Name())

		// 读取并解析JSON文件
		images, err := v.ReadJsonFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("解析文件 %s 失败: %v", file.Name(), err)
		}

		// 合并结果
		result = append(result, images...)
	}

	return result, nil
}

// 获取所有漏洞镜像信息
func (v *VulService) GetVulImages() (VulImagesList, error) {
	result := VulImagesList{}
	// 从本地获取漏洞镜像信息
	LocalVulImagesList, err := v.GetVulImageFromLocal(config.LocalImagePath)
	if err != nil {
		return nil, err
	}
	result = append(result, LocalVulImagesList...)
	return result, nil
}

// 获取本地镜像仓库地址
func (v *VulService) GetVulStoragePath() string {
	return config.LocalImagePath
}

// 修改本地镜像仓库地址
func (v *VulService) SetVulStoragePath(path string) error {
	config.LocalImagePath = path
	return nil
}

// 保存上传的JSON文件到本地镜像仓库
func (v *VulService) SaveUploadedJsonFile(filename string, base64FileData string) error {
	// 检查文件扩展名
	if !strings.HasSuffix(filename, ".json") {
		return fmt.Errorf("只支持.json格式的文件")
	}

	// 解码Base64数据
	fileData, err := base64.StdEncoding.DecodeString(base64FileData)
	if err != nil {
		return fmt.Errorf("Base64解码失败: %v", err)
	}

	// 验证JSON结构
	var imagesList VulImagesList
	if err = json.Unmarshal(fileData, &imagesList); err == nil {
		// 验证每个VulImage的结构
		for _, img := range imagesList {
			if img.ImageName == "" || img.ImageVulName == "" {
				return utils.ErrJsonMarshal
			}
		}
	} else {
		// 尝试解析为单个VulImage
		var image VulImage
		if err = json.Unmarshal(fileData, &image); err != nil {
			return utils.ErrJsonMarshal
		}
		if image.ImageName == "" || image.ImageVulName == "" {
			return utils.ErrJsonMarshal
		}
	}
	// 构建完整保存路径
	cleanFilename := filepath.Clean(filename)
	if strings.Contains(cleanFilename, "..") || filepath.IsAbs(cleanFilename) {
		middleware.SugarLogger.Errorf("非法文件名: %s", filename)
		return fmt.Errorf("非法文件名: %s", filename)
	}

	// 防御路径穿越攻击
	if strings.Contains(cleanFilename, string(filepath.Separator)) {
		middleware.SugarLogger.Errorf("文件名包含路径分隔符: %s", filename)
		return fmt.Errorf("文件名包含路径分隔符")
	}

	filePath := filepath.Join(config.LocalImagePath, cleanFilename)

	// 二次校验最终路径是否在目标目录下（防御编码攻击）
	if !strings.HasPrefix(filepath.Clean(filePath), filepath.Clean(config.LocalImagePath)) {
		middleware.SugarLogger.Errorf("非法存储路径: %s", filePath)
		return fmt.Errorf("非法存储路径: %s", filePath)
	}

	// 使用utils.SaveFile保存文件
	if err := utils.SaveFile(fileData, filePath); err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	return nil
}

type VulEnvList []VulEnv

type VulEnv struct {
	ID           uint      `json:"-"`
	EnvName      string    `json:"env_name"`
	EnvDesc      string    `json:"env_desc"`
	EnvType      string    `json:"env_type"`
	Base_Image   string    `json:"base_image"`
	Base_compose string    `json:"base_compose"`
	Rank         float64   `json:"rank"`
	From         string    `json:"from"`
	Degree       VulDegree `json:"degree"`
	Cost         float64   `json:"cost"`
	IsOpen       int       `json:"is_open"`
}

// 将model.VulEnv转换为VulEnv
func ConvertToVulEnv(vulEnv *model.VulEnv) *VulEnv {
	if vulEnv == nil {
		return nil
	}
	// 转换VulDegree字段
	var degree VulDegree
	if err := json.Unmarshal([]byte(vulEnv.Degree), &degree); err != nil {
		middleware.SugarLogger.Errorf("解析VulDegree字段失败: %v", err)
	}
	return &VulEnv{
		ID:           vulEnv.ID,
		EnvName:      vulEnv.EnvName,
		EnvDesc:      vulEnv.EnvDesc,
		EnvType:      vulEnv.EnvType,
		Base_Image:   vulEnv.BaseImage,
		Base_compose: vulEnv.BaseCompose,
		Rank:         vulEnv.Rank,
		From:         vulEnv.From,
		Degree:       degree,
		Cost:         vulEnv.Cost,
		IsOpen:       vulEnv.IsOpen,
	}

}

// 获取所有漏洞环境信息
func (v *VulService) GetVulEnv() (VulEnvList, error) {
	// 获取Docker本地镜像列表
	dockerImages, err := GetImages()
	if err != nil {
		return nil, err
	}

	// 一次性获取所有漏洞镜像数据（避免在循环中重复获取）
	VulSavedData, err := v.GetVulImages()
	if err != nil {
		return nil, err
	}

	var result VulEnvList
	for _, image := range dockerImages {
		var vulEnv VulEnv
		var found bool

		// 查找对应的漏洞镜像
		for _, vulImage := range VulSavedData {
			if vulImage.ImageName == image.Name {
				// 使用已有配置信息
				vulEnv = VulEnv{
					EnvName:    vulImage.ImageVulName,
					EnvDesc:    vulImage.ImageDesc,
					EnvType:    "单镜像",
					Base_Image: image.Name,
					Rank:       vulImage.Rank,
					From:       vulImage.From,
					Degree:     vulImage.Degree,
				}
				found = true
				break
			}
		}

		if !found {
			vulEnv = VulEnv{
				EnvName:    image.Name,
				EnvDesc:    "暂无信息",
				EnvType:    "单镜像",
				Base_Image: image.Name,
				Rank:       3.5, // 默认评分
				From:       "单镜像",
				Degree:     VulDegree{},
			}
		}
		result = append(result, vulEnv)
	}

	// 获取Docker Compose文件列表
	composeFiles, err := v.GetDockerComposeFiles()
	if err != nil {
		middleware.SugarLogger.Errorf("获取Docker Compose文件失败: %v", err)
	} else {
		// 处理每个compose文件
		for _, filePath := range composeFiles {
			// 获取目录名作为环境名称
			dirName := filepath.Base(filepath.Dir(filePath))

			result = append(result, VulEnv{
				EnvName:      dirName,
				EnvDesc:      "基于Docker Compose的多服务环境",
				EnvType:      "复合环境",
				Base_compose: filePath,
				Rank:         4.0, // 默认评分高于单镜像
				From:         "compose",
				Degree:       VulDegree{},
			})
		}
	}

	return result, nil
}

// GetDockerComposeFiles 遍历目录获取所有docker-compose.yml文件
func (v *VulService) GetDockerComposeFiles() ([]string, error) {
	var composeFiles []string

	// 使用filepath.Walk遍历目录
	err := filepath.Walk(config.LocalImagePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 匹配文件名且非目录
		if !info.IsDir() && strings.HasSuffix(info.Name(), "docker-compose.yml") {
			composeFiles = append(composeFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历目录失败: %v", err)
	}

	return composeFiles, nil
}

// 上传docker compose 压缩包
func (v *VulService) UploadVulZip(base64Data string) error {
	// 解码Base64数据
	decodedData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		middleware.SugarLogger.Errorf("Base64解码失败: %v", err)
		return err
	}

	// 解压文件到指定目录
	err = utils.Unzip(decodedData, config.LocalImagePath)
	if err != nil {
		middleware.SugarLogger.Errorf("解压文件失败: %v", err)
		return err
	}
	return nil
}

// 创建漏洞环境
func (v *VulService) CreateVulEnv(vulEnv *VulEnv, conn *websocket.Conn) error {

	if vulEnv.EnvName == "" || (vulEnv.Base_Image == "" && vulEnv.Base_compose == "") {
		return fmt.Errorf("缺少必要字段")
	}

	// 检查环境名称是否已存在
	if _, err := model.GetVulEnvByName(vulEnv.EnvName); err == nil {
		return fmt.Errorf("环境名称已存在")
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go MonitorStopOperation(conn, cancel)
	// 如果Base_Image存在 检查镜像是否存在
	if vulEnv.Base_Image != "" {
		if ok := ImageExists(vulEnv.Base_Image); !ok {
			// 开始拉取镜像
			if err := PullImage(ctx, vulEnv.Base_Image, conn); err != nil {
				return fmt.Errorf("拉取镜像失败: %v", err)
			}
		}
	}

	// 如果Base_compose存在 检查文件是否存在
	if vulEnv.Base_compose != "" {
		if strings.Compare(vulEnv.Base_compose[len(vulEnv.Base_compose)-4:], ".yml") != 0 {
			return fmt.Errorf("错误的文件路径")
		}
		if ok := utils.IsPathExist(vulEnv.Base_compose); !ok {
			return fmt.Errorf("文件不存在")
		}
		// 读取需要拉取的镜像
		images, err := GetImagesFromCompose(vulEnv.Base_compose)
		if err != nil {
			return fmt.Errorf("读取docker-compose.yml失败: %v", err)
		}
		// 开始拉取镜像
		for _, image := range images {
			if ok := ImageExists(image); !ok {
				if err := PullImage(ctx, image, conn); err != nil {
					return err
				}
			}
		}
	}

	// 将degree序列化为JSON字符串
	degreeJSON, err := json.Marshal(vulEnv.Degree)
	if err != nil {
		return fmt.Errorf("JSON序列化失败: %v", err)
	}

	// 转换service层结构体到model层结构体
	newVul := model.VulEnv{
		EnvName:     vulEnv.EnvName,
		EnvDesc:     vulEnv.EnvDesc,
		EnvType:     vulEnv.EnvType,
		BaseImage:   vulEnv.Base_Image,
		BaseCompose: vulEnv.Base_compose,
		Rank:        vulEnv.Rank,
		From:        vulEnv.From,
		Degree:      string(degreeJSON),
		IsOpen:      vulEnv.IsOpen,
		Cost:        vulEnv.Cost,
	}

	// 调用model层方法
	if err := model.CreateVulEnv(&newVul); err != nil {
		return fmt.Errorf("创建失败: %v", err)
	}
	return nil
}

// 删除漏洞环境
func (v *VulService) DeleteVulEnv(EnvID uint, isDeleteImage bool) error {
	// 先删除实例
	err := v.StopVulInstanceByVulEnvID(EnvID)
	if err != nil {
		return err
	}

	// 删除实例数据表记录
	if err := model.DeleteVulInstanceByVulEnvID(EnvID); err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	// 如果需要删除镜像 先获取漏洞环境
	env, err := model.GetVulEnvByID(EnvID)
	if err != nil {
		return err
	}

	if isDeleteImage {
		if env.BaseCompose != "" {
			// 删除镜像
			if err = DeleteComposeImages(env.BaseCompose); err != nil {
				return err
			}
		} else {
			if err = DeleteImage(env.BaseImage); err != nil {
				return err
			}
		}
	}

	// 删除漏洞环境数据表记录
	if err := model.DeleteVulEnv(EnvID); err != nil {
		return err
	}
	return nil
}

type VulInstanceList []VulInstanceService

// VulInstanceService 用户漏洞环境实例服务层结构体
type VulInstanceService struct {
	VulEnv
	UserDTO
	ID          uint      `json:"id"`
	UserID      uint      `json:"user_id"`
	VulEnvID    uint      `json:"vul_env_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      int       `json:"status"` // 1-运行中 2-已停止 3-已完成
	ContainerID string    `json:"container_id,omitempty"`
	StackName   string    `json:"stack_name,omitempty"`
	Ports       []string  `json:"ports"` // 字符串转为数组方便前端使用
	ExpireTime  time.Time `json:"expire_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// 转换模型到服务层结构
func ConvertVulInstanceModelToService(model *model.VulInstance) *VulInstanceService {
	// 解析端口字符串为数组
	var ports []string
	if model.Ports != "" {
		ports = strings.Split(model.Ports, ",")
	}
	result := &VulInstanceService{
		ID:          model.ID,
		UserID:      model.UserID,
		VulEnvID:    model.VulEnvID,
		StartTime:   model.StartTime,
		EndTime:     model.EndTime,
		Status:      model.Status,
		ContainerID: model.ContainerID,
		StackName:   model.StackName,
		Ports:       ports,
		ExpireTime:  model.ExpireTime,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
	result.VulEnv = VulEnv{
		EnvName:      model.VulEnv.EnvName,
		EnvDesc:      model.VulEnv.EnvDesc,
		EnvType:      model.VulEnv.EnvType,
		Base_Image:   model.VulEnv.BaseImage,
		Base_compose: model.VulEnv.BaseCompose,
		Rank:         model.VulEnv.Rank,
		From:         model.VulEnv.From,
		Cost:         model.VulEnv.Cost,
		IsOpen:       model.VulEnv.IsOpen,
	}
	json.Unmarshal([]byte(model.VulEnv.Degree), &result.VulEnv.Degree) // 反序列化degree字段
	return result
}

// 获取所有已经创建的漏洞环境
func (v *VulService) GetVulEnvList(roleLevel int) (VulEnvList, error) {
	// 调用model层方法
	vulEnvList, err := model.GetVulEnvsByOpenLevel(roleLevel)
	if err != nil {
		return nil, fmt.Errorf("获取失败: %v", err)
	}
	result := VulEnvList{}
	for _, vul := range vulEnvList {
		// 反序列化degree字段
		var degree VulDegree
		if err := json.Unmarshal([]byte(vul.Degree), &degree); err != nil {
			middleware.SugarLogger.Errorf("JSON反序列化失败: %v", err)
			continue // 跳过反序列化失败的环境
		}

		// 构建VulImage结构体
		vulImage := VulEnv{
			ID:           vul.ID,
			EnvName:      vul.EnvName,
			EnvDesc:      vul.EnvDesc,
			EnvType:      vul.EnvType,
			Base_Image:   vul.BaseImage,
			Base_compose: vul.BaseCompose,
			Rank:         vul.Rank,
			From:         vul.From,
			Degree:       degree,
			Cost:         vul.Cost,
			IsOpen:       vul.IsOpen,
		}
		result = append(result, vulImage)
	}
	return result, nil
}

// 获取所有用户开启的漏洞环境
func (v *VulService) GetVulInstanceList() (VulInstanceList, error) {
	// 调用model层方法
	vulInstanceList, err := model.GetAllVulInstances()
	if err != nil {
		return nil, fmt.Errorf("获取失败: %v", err)
	}
	result := VulInstanceList{}
	for _, vul := range vulInstanceList {
		// 反序列化degree字段
		var degree VulDegree
		if err := json.Unmarshal([]byte(vul.VulEnv.Degree), &degree); err != nil {
			middleware.SugarLogger.Errorf("JSON反序列化失败: %v", err)
			continue // 跳过反序列化失败的环境
		}
	}
	return result, nil
}

// 创建场景实例
func (v *VulService) CreateVulInstance(userID uint, vulEnvID uint) (*VulInstanceService, error) {
	// 检查环境是否存在
	VulEnv, err := model.GetVulEnvByID(vulEnvID)
	if err != nil {
		return nil, fmt.Errorf("环境不存在")
	}

	// 检查用户是否存在
	user, err := model.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 检查镜像是否对用户开放
	if VulEnv.IsOpen < RoleMap[user.Role] {
		return nil, fmt.Errorf("镜像未开放")
	}

	// 检查用户是否有足够的余额
	if user.Score < VulEnv.Cost {
		return nil, fmt.Errorf("余额不足")
	}

	// 扣除用户余额
	if err = model.UpdateUserScore(userID, user.Score-VulEnv.Cost); err != nil {
		return nil, fmt.Errorf("扣除余额失败: %v", err)
	}

	// 检查用户是否已经有该环境的实例
	instance, err := model.GetVulInstanceBy2ID(userID, vulEnvID)
	if err == nil && instance.Status == 1 { // 1 表示运行中
		return nil, fmt.Errorf("用户已经有该环境的实例")
	}

	// 创建场景实例
	newVulInstance := model.VulInstance{}

	ports := map[string]string{}

	// 检查需要的镜像是否存在并开启环境
	if VulEnv.BaseImage != "" {
		// 检查镜像是否存在
		if ok := ImageExists(VulEnv.BaseImage); !ok {
			return nil, fmt.Errorf("镜像不存在")
		}
		// 获取镜像端口映射
		ports, err = GeneratePortBindings(VulEnv.BaseImage)
		if err != nil {
			return nil, fmt.Errorf("获取镜像端口映射失败: %v", err)
		}
		// 启动镜像
		containerName := normalizeProjectName(utils.MD5Encode(string(user.ID) + VulEnv.EnvName))
		containerID, err := CreateContainer(VulEnv.BaseImage, containerName, nil, ports)
		if err != nil {
			return nil, fmt.Errorf("启动镜像失败: %v", err)
		}
		newVulInstance.ContainerID = containerID
	}
	if VulEnv.BaseCompose != "" {
		imageList, err := GetImagesFromCompose(VulEnv.BaseCompose)
		if err != nil {
			return nil, fmt.Errorf("读取docker-compose.yml失败: %v", err)
		}
		for _, image := range imageList {
			if ok := ImageExists(image); !ok {
				return nil, fmt.Errorf("镜像不存在")
			}
		}
		// 启动docker compose 环境
		ports = map[string]string{}
		stackName := normalizeProjectName(utils.MD5Encode(string(user.ID) + VulEnv.EnvName))
		err = CreateFromCompose(VulEnv.BaseCompose, stackName, &ports)
		if err != nil {
			RemoveStackByName(stackName)
			return nil, fmt.Errorf("启动docker compose 环境失败: %v", err)
		}
		newVulInstance.StackName = stackName
	}
	portsStr, err := json.Marshal(ports)
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %v", err)
	}
	newVulInstance.Ports = string(portsStr)
	// 填充基本信息
	newVulInstance.UserID = userID
	newVulInstance.VulEnvID = vulEnvID
	newVulInstance.Status = 1 // 1 表示运行中
	newVulInstance.StartTime = time.Now()
	newVulInstance.ExpireTime = time.Now().Add(config.DefaultExpirationTime)
	// 调用model层方法
	if err := model.CreateVulInstance(&newVulInstance); err != nil {
		return nil, fmt.Errorf("创建失败: %v", err)
	}
	result := ConvertVulInstanceModelToService(&newVulInstance)
	return result, nil
}

// 获取指定用户的漏洞实例
func (v *VulService) GetVulInstanceByUserID(userID uint) (VulInstanceList, error) {
	// 调用model层方法
	vulInstanceList, err := model.GetVulInstanceByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("获取失败: %v", err)
	}
	result := VulInstanceList{}
	for _, vul := range vulInstanceList {
		result = append(result, *ConvertVulInstanceModelToService(&vul))
	}
	return result, nil
}

// 根据vulEnvID删除实例环境
func (v *VulService) StopVulInstanceByVulEnvID(vulEnvID uint) error {
	// 获取实例信息
	instances, err := model.GetVulInstanceByVulEnvID(vulEnvID)
	if err != nil {
		return err
	}

	for _, instance := range instances {
		// 根据实例类型执行不同的删除逻辑
		if instance.ContainerID != "" {
			// 删除单容器实例
			if err := RemoveContainer(instance.ContainerID, true); err != nil {
				return fmt.Errorf("删除容器失败: %v", err)
			}
		} else if instance.StackName != "" {
			// 删除Compose堆栈实例
			if err := RemoveStackByName(instance.StackName); err != nil {
				return fmt.Errorf("删除堆栈失败: %v", err)
			}
		}
	}
	return nil
}

// 删除指定用户的实例环境
func (v *VulService) DeleteVulInstance(userID uint, vulEnvID uint) error {
	// 获取实例信息
	instance, err := model.GetVulInstanceBy2ID(userID, vulEnvID)
	if err != nil {
		return fmt.Errorf("实例不存在")
	}

	// 根据实例类型执行不同的删除逻辑
	if instance.ContainerID != "" {
		// 删除单容器实例
		if err := RemoveContainer(instance.ContainerID, true); err != nil {
			return fmt.Errorf("删除容器失败: %v", err)
		}
	} else if instance.StackName != "" {
		// 删除Compose堆栈实例
		if err := RemoveStackByName(instance.StackName); err != nil {
			return fmt.Errorf("删除堆栈失败: %v", err)
		}
	}

	// 更新数据库状态为已删除
	if err := model.DeleteVulInstanceBy2ID(userID, vulEnvID); err != nil {
		return fmt.Errorf("更新数据库失败: %v", err)
	}

	return nil
}

// 获取所有实例
func (v *VulService) GetAllVulInstances() (VulInstanceList, error) {
	// 调用model层方法
	vulInstanceList, err := model.GetAllVulInstances()
	if err != nil {
		return nil, fmt.Errorf("获取失败: %v", err)
	}
	result := VulInstanceList{}
	for _, vul := range vulInstanceList {
		// 添加用户信息
		u, err := model.GetUserByID(vul.UserID)
		if err != nil {
			return nil, fmt.Errorf("获取用户信息失败: %v", err)
		}
		instance := *ConvertVulInstanceModelToService(&vul)
		instance.UserDTO = invertUserDTO(*u)
		mvulEnv, err := model.GetVulEnvByID(vul.VulEnvID)
		if err != nil {
			return nil, fmt.Errorf("获取漏洞环境信息失败: %v", err)
		}
		instance.VulEnv = *ConvertToVulEnv(mvulEnv)
		result = append(result, instance)
	}
	return result, nil
}

// 启动定时监控过期实例
func StartMonitorExpiredInstances() {
	ticker := time.NewTicker(1 * time.Minute)
	v := &VulService{}
	go func() {
		for range ticker.C {
			v.checkAndCleanExpiredInstances()
		}
	}()
}

// 检查并清理过期实例
func (v *VulService) checkAndCleanExpiredInstances() error {
	// 获取所有运行中的实例
	instances, err := model.GetAllVulInstances()
	if err != nil {
		middleware.SugarLogger.Errorf("监控器获取实例列表失败: %v", err)
		return err
	}

	now := time.Now()
	for _, instance := range instances {
		// 检查实例是否已过期且仍在运行
		if instance.ExpireTime.Before(now) {
			err = v.DeleteVulInstance(instance.UserID, instance.VulEnvID)
			if err != nil {
				middleware.SugarLogger.Errorf("监控器删除实例失败: %v", err)
				return err
			}
		}
	}
	return nil
}

// 延长实例时间
func (v *VulService) ExtendExpireTime(id uint) error {
	return model.ExtendExpireTime(id)
}
