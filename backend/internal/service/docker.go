package service

import (
	"AscensionPath/config"
	"AscensionPath/internal/middleware"
	"AscensionPath/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"path/filepath"

	"github.com/compose-spec/compose-go/v2/loader"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gorilla/websocket"
)

const (
	maxImageNameLength = 128
	defaultTimeout     = 10 * time.Minute
)

// 创建 Docker 客户端 (复用代码)
var (
	dockerCli *client.Client
	cliOnce   sync.Once
)

func init() {
	// 初始化 Docker 客户端
	initDockerClient()
}

// 初始化Docker客户端
func initDockerClient() {
	var err error
	dockerCli, err = client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithTimeout(defaultTimeout),
	)
	if err != nil {
		middleware.SugarLogger.Errorf("初始化Docker客户端失败: %v", err)
	}
}

// 检查Docker客户端是否可用
func IsDockerAvailable() (bool, error) {
	cli := dockerCli

	// 执行一个简单的ping操作验证连接
	_, err := cli.Ping(context.Background())
	if err != nil {
		middleware.SugarLogger.Errorf("无法连接到Docker守护进程: %v", err)
		return false, err
	}

	return true, nil
}

// 获取Docker客户端 (线程安全)
func getDockerClient() (*client.Client, error) {
	if ok, err := IsDockerAvailable(); !ok {
		initDockerClient()
		return dockerCli, err
	}
	return dockerCli, nil
}

// 定义镜像信息结构体
type Image struct {
	ID        string    `json:"id"`        // 镜像ID(短格式)
	Name      string    `json:"name"`      // 镜像名称
	Tags      []string  `json:"tags"`      // 镜像标签列表
	SizeMB    float64   `json:"sizeMB"`    // 镜像大小(MB)
	CreatedAt time.Time `json:"createdAt"` // 创建时间
}

// 定义镜像列表类型
type ImagesList []Image

// 从本地获取镜像列表
func GetImages() (ImagesList, error) {
	cli, err := getDockerClient()
	if err != nil {
		return ImagesList{}, err
	}

	images, err := cli.ImageList(context.Background(), image.ListOptions{All: true})
	if err != nil {
		middleware.SugarLogger.Errorf("获取镜像列表失败: %v", err)
		return ImagesList{}, err
	}

	result := make(ImagesList, 0, len(images))
	for _, img := range images {
		result = append(result, Image{
			ID:        shortenID(img.ID),
			Name:      getImageName(img.RepoTags),
			Tags:      img.RepoTags,
			SizeMB:    float64(img.Size) / (1024 * 1024),
			CreatedAt: time.Unix(img.Created, 0),
		})
	}

	return result, nil
}

// 辅助函数
func getImageName(repoTags []string) string {
	if len(repoTags) == 0 {
		return "<untagged>"
	}
	return repoTags[0]
}

func shortenID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// 定义 convertMappingToSlice 函数，将 types.MappingWithEquals 类型转换为 []string 类型
func convertMappingToSlice(mapping types.MappingWithEquals) []string {
	var result []string
	for key, value := range mapping {
		if value != nil {
			// 保留键值对格式 key=value
			result = append(result, fmt.Sprintf("%s=%s", key, *value))
		} else {
			// 如果值为nil，只保留键名
			result = append(result, key)
		}
	}
	return result
}

// 检查镜像是否存在
func ImageExists(imageName string) bool {
	cli, err := getDockerClient()
	if err != nil {
		return false
	}
	_, err = cli.ImageInspect(context.Background(), imageName)
	return err == nil
}

// MonitorStopOperation 回调函数 监控停止操作
func MonitorStopOperation(conn *websocket.Conn, fn context.CancelFunc) {
	if conn != nil {
		// 读取具体消息内容
		msg, _ := utils.ReadWSMessage[struct{ Action string }](conn)
		// 处理特定action
		if msg.Data.Action == "CANCEL_PULL" {
			middleware.SugarLogger.Infof("接收到手动终止指令")
			fn()
			return
		}
	}
}

// PullImage 拉取Docker镜像(支持代理)
func PullImage(ctx context.Context, imageName string, conn *websocket.Conn) error {
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	// 创建带代理选项的PullOptions
	pullOpts := image.PullOptions{}

	// 从环境变量获取代理设置
	if proxyURL := config.Proxy; proxyURL != "" {
		pullOpts.RegistryAuth = "" // 如果需要认证可以在这里设置
		// 设置代理环境变量
		err = os.Setenv("HTTP_PROXY", proxyURL)
		if err != nil {
			return fmt.Errorf("设置HTTP_PROXY环境变量失败: %v", err)
		}
		os.Setenv("HTTPS_PROXY", proxyURL)
		defer func() {
			// 清理代理设置
			os.Unsetenv("HTTP_PROXY")
			os.Unsetenv("HTTPS_PROXY")
		}()
	}

	// 先检查本地是否已存在该镜像
	_, err = cli.ImageInspect(context.Background(), imageName)
	if err == nil {
		re := fmt.Sprintf(" 本地已存在镜像 %s，跳过拉取", imageName)
		middleware.SugarLogger.Infof(re)
		return errors.New(re)
	} else if !client.IsErrNotFound(err) {
		return fmt.Errorf("检查本地镜像失败: %v", err)
	}

	// 拉取镜像
	out, err := cli.ImagePull(ctx, imageName, pullOpts)
	if err != nil {
		return fmt.Errorf("拉取镜像失败: %v", err)
	}
	defer out.Close()

	// 实时解析并发送进度
	decoder := json.NewDecoder(out)
	for {
		var progress struct {
			Status   string `json:"status"`
			Progress string `json:"progress"`
			ID       string `json:"id"`
		}

		if err = decoder.Decode(&progress); err != nil {
			if ctx.Err() != nil { // 检查是否是上下文取消导致的错误
				return fmt.Errorf("拉取 %s 镜像被取消", imageName) // 返回自定义错误，不包含原始错误信息
			}
			if err == io.EOF {
				break
			}
			middleware.SugarLogger.Errorf("进度解析失败: %v", err)
			// 检查是否拉取成功(函数终止位置)
			if ImageExists(imageName) {
				middleware.SugarLogger.Infof("成功拉取镜像: %s", imageName)
				return nil
			} else {
				return fmt.Errorf("拉取 %s 镜像失败", imageName)
			}
		}

		// 通过WebSocket发送进度
		if conn != nil {
			msg := utils.Message[any]{
				Code:    utils.CodeSuccess,
				Message: progress.Status,
				Data:    progress,
			}
			if err = conn.WriteJSON(msg); err != nil {
				middleware.SugarLogger.Errorf("发送进度失败: %v", err)
				break
			}
		}
	}
	// 进度发送失败后检查是否拉取成功
	if ImageExists(imageName) {
		middleware.SugarLogger.Infof("成功拉取镜像: %s", imageName)
		return nil
	} else {
		return fmt.Errorf("拉取 %s 镜像失败", imageName)
	}
}

// 新增函数：从docker compose文件中解析需要拉取的镜像列表
func GetImagesFromCompose(composePath string) ([]string, error) {
	// 读取compose文件内容
	composeData, err := os.ReadFile(composePath)
	if err != nil {
		middleware.SugarLogger.Errorf("读取 compose 文件失败: %v", err)
		return nil, fmt.Errorf("读取 compose 文件失败: %v", err)
	}

	// 规范化项目名称
	rawName := filepath.Base(filepath.Dir(composePath))
	stackName := normalizeProjectName(rawName)

	// 解析compose文件
	project, err := loader.LoadWithContext(context.Background(),
		types.ConfigDetails{
			ConfigFiles: []types.ConfigFile{
				{
					Filename: composePath,
					Content:  composeData,
				},
			},
			Environment: map[string]string{
				"COMPOSE_PROJECT_NAME": stackName, // 使用规范化后的名称
			},
		},
		func(o *loader.Options) {
			o.ResolvePaths = true
			o.SetProjectName(stackName, true) // 显式设置项目名称
		},
	)
	if err != nil {
		middleware.SugarLogger.Errorf("解析 compose 文件失败: %v", err)
		return nil, fmt.Errorf("解析 compose 文件失败: %v", err)
	}

	// 收集所有服务使用的镜像
	var images []string
	for _, service := range project.Services {
		if service.Image != "" {
			images = append(images, service.Image)
		}
	}

	return images, nil
}

// 使用 compose-go 解析并部署 Docker Compose 文件
func CreateFromCompose(composePath, stackName string, ports *map[string]string) error {

	labels := map[string]string{
		"com.docker.compose.project": stackName,
		"com.docker.compose.oneoff":  "False",
	}

	// 创建专用网络（使用规范化名称）
	networkName := fmt.Sprintf("%s_default", stackName)
	networkID, err := createNetworkWithLabels(networkName, labels)
	if err != nil {
		return err
	}

	// 读取并解析compose文件
	composeData, err := os.ReadFile(composePath)
	if err != nil {
		middleware.SugarLogger.Errorf("读取 compose 文件失败: %v", err)
		return err
	}

	// 修改解析compose文件部分
	project, err := loader.LoadWithContext(context.Background(),
		types.ConfigDetails{
			ConfigFiles: []types.ConfigFile{
				{
					Filename: composePath,
					Content:  composeData,
				},
			},
			Environment: map[string]string{
				"COMPOSE_PROJECT_NAME": stackName, // 使用规范化后的名称
			},
		},
		func(o *loader.Options) {
			o.SetProjectName(stackName, true) // 使用规范化后的名称
			o.ResolvePaths = true
		},
	)
	if err != nil {
		middleware.SugarLogger.Errorf("解析 compose 文件失败: %v", err)
		return err
	}

	// 先部署无依赖的服务
	for _, service := range project.Services {
		if len(service.DependsOn) == 0 {
			if err := deployService(service, networkID, composePath, stackName, ports); err != nil {
				return err
			}
		}
	}
	// 按依赖层级部署服务
	for {
		deployed := 0
		for _, service := range project.Services {
			if len(service.DependsOn) > 0 {
				// 检查所有依赖是否已部署
				allDepsReady := true
				// 检查依赖服务是否已部署
				for dep := range service.DependsOn {
					if !isServiceDeployed(project, dep) {
						// 依赖服务未部署
						allDepsReady = false
						break
					}
				}

				if allDepsReady {
					if !isServiceDeployed(project, service.Name) {
						if err := deployService(service, networkID, composePath, stackName, ports); err != nil {
							return err
						}
						deployed++
					}
				}
			}
		}

		// 如果没有服务被部署，说明有循环依赖或错误
		if deployed == 0 {
			break
		}
	}

	// 检查是否有未部署的服务
	for _, service := range project.Services {
		if !isServiceDeployed(project, service.Name) {
			return fmt.Errorf("服务 %s 无法部署，可能存在循环依赖或未满足的依赖", service.Name)
		}
	}

	middleware.SugarLogger.Infof("成功部署堆栈 %s 包含 %d 个服务", stackName, len(project.Services))
	return nil
}

// 新增带标签的网络创建函数
func createNetworkWithLabels(name string, labels map[string]string) (string, error) {
	cli, err := getDockerClient()
	if err != nil {
		return "", err
	}

	resp, err := cli.NetworkCreate(context.Background(), name, network.CreateOptions{
		Driver: "bridge",
		Labels: labels,
		IPAM: &network.IPAM{
			Driver: "default",
		},
	})
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

// PortMapToPortSet 将 nat.PortMap 转换为 nat.PortSet
func PortMapToPortSet(portMap nat.PortMap) nat.PortSet {
	portSet := make(nat.PortSet)
	for port := range portMap {
		portSet[port] = struct{}{}
	}
	return portSet
}

// deployService 根据 compose 文件创建容器
func deployService(service types.ServiceConfig, networkID string, composePath string, stackName string, ports *map[string]string) error {
	// 检查并拉取镜像
	cli, err := getDockerClient()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err = cli.ImageInspect(context.Background(), service.Image)
	if err != nil {
		if client.IsErrNotFound(err) {
			middleware.SugarLogger.Infof("镜像 %s 不存在，开始拉取...", service.Image)
			if err = PullImage(ctx, service.Image, nil); err != nil {
				middleware.SugarLogger.Errorf("拉取镜像 %s 失败: %v", service.Image, err)
				return err
			}
		} else {
			middleware.SugarLogger.Errorf("检查镜像 %s 失败: %v", service.Image, err)
			return err
		}
	}

	// 构建容器配置（添加堆栈标签）
	containerConfig := &container.Config{
		Image: service.Image,
		Env:   convertMappingToSlice(service.Environment),
		Labels: map[string]string{
			"com.docker.compose.project": stackName, // 使用规范化后的名称
			"com.docker.compose.service": service.Name,
			"com.docker.compose.oneoff":  "False",
		},
	}

	// 添加command配置
	if service.Command != nil {
		// 将 types.ShellCommand 类型的 service.Command 转换为 []string 类型
		var cmdSlice []string
		for _, cmd := range service.Command {
			cmdSlice = append(cmdSlice, string(cmd))
		}
		containerConfig.Cmd = cmdSlice
	}

	// 构建主机配置
	hostConfig := &container.HostConfig{}

	// 添加卷挂载
	if len(service.Volumes) > 0 {
		mounts := make([]mount.Mount, 0, len(service.Volumes))
		for _, vol := range service.Volumes {
			// 处理相对路径，转换为绝对路径
			source := vol.Source
			if !filepath.IsAbs(source) {
				source = filepath.Join(filepath.Dir(composePath), source)
			}
			mounts = append(mounts, mount.Mount{
				Type:   mount.TypeBind,
				Source: source,
				Target: vol.Target,
			})
		}
		hostConfig.Mounts = mounts
	}

	if len(service.Ports) > 0 {
		portBindings := make(nat.PortMap)
		for _, portSpec := range service.Ports {
			if portSpec.Published != "" {
				containerPort := nat.Port(fmt.Sprintf("%d/%s", portSpec.Target, portSpec.Protocol))
				// 动态获取可用主机端口
				availablePort, err := utils.GetAvailablePort()
				if err != nil {
					middleware.SugarLogger.Errorf("获取可用端口失败: %v", err)
					return err
				}
				portBindings[containerPort] = []nat.PortBinding{
					{
						HostPort: strconv.Itoa(availablePort), // 使用动态获取的端口
					},
				}
			}
		}
		hostConfig.PortBindings = portBindings
		containerConfig.ExposedPorts = PortMapToPortSet(portBindings)
		// 将端口映射赋值给函数参数
		if ports != nil {
			for containerPort, bindings := range portBindings {
				if len(bindings) > 0 {
					(*ports)[containerPort.Port()] = bindings[0].HostPort
				}
			}
		}
	}

	// 创建并启动容器（确保网络配置包含堆栈信息）
	networkingConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			networkID: {
				Aliases: []string{service.Name},
				// 添加堆栈标签到网络端点
				DriverOpts: map[string]string{
					"com.docker.compose.project": filepath.Base(filepath.Dir(networkID)),
				},
			},
		},
	}

	resp, err := cli.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		networkingConfig, // 使用包含别名的网络配置
		nil,
		service.Name,
	)
	if err != nil {
		middleware.SugarLogger.Errorf("创建容器 %s 失败: %v", service.Name, err)
		return err
	}

	// 先连接容器到网络
	if err := cli.NetworkConnect(context.Background(), networkID, resp.ID, nil); err != nil {
		middleware.SugarLogger.Errorf("连接容器 %s 到网络 %s 失败: %v", resp.ID, networkID, err)
		return err
	}

	// 再启动容器
	if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		middleware.SugarLogger.Errorf("启动容器 %s 失败: %v", service.Name, err)
		return err
	}

	middleware.SugarLogger.Infof("成功部署并启动服务: %s (ID: %s)", service.Name, resp.ID)
	return nil
}

// 获取当前运行的所有容器信息
func GetRunningContainers() ([]container.Summary, error) {
	cli, err := getDockerClient()
	if err != nil {
		return nil, err
	}

	// 获取容器列表，只包含运行中的容器
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: false, // 只返回运行中的容器
	})
	if err != nil {
		middleware.SugarLogger.Errorf("获取容器列表失败: %v", err)
		return nil, err
	}

	// 记录获取到的容器数量
	middleware.SugarLogger.Infof("获取到 %d 个运行中的容器", len(containers))

	return containers, nil
}

// 创建并启动容器
// 全局变量保存创建的容器信息
var (
	createdContainers = make(map[string]string) // key: containerID, value: containerName
	containersMutex   sync.RWMutex
)

// 获取已经创建的容器信息
func GetContainerInfo() map[string]string {
	return createdContainers
}

// GetImageExposedPorts 获取镜像暴露的端口
func GetImageExposedPorts(imageName string) (nat.PortSet, error) {
	cli, err := getDockerClient()
	if err != nil {
		return nil, err
	}

	inspect, err := cli.ImageInspect(context.Background(), imageName)
	if err != nil {
		return nil, err
	}

	return inspect.Config.ExposedPorts, nil
}

// GeneratePortBindings 生成端口绑定映射
func GeneratePortBindings(imageName string) (map[string]string, error) {
	// 获取镜像暴露的端口
	exposedPorts, err := GetImageExposedPorts(imageName)
	if err != nil {
		return nil, err
	}

	bindings := make(map[string]string)

	// 为每个暴露的端口分配一个可用端口
	for port := range exposedPorts {
		availablePort, err := utils.GetAvailablePort()
		if err != nil {
			return nil, fmt.Errorf("获取可用端口失败: %v", err)
		}
		bindings[port.Port()] = strconv.Itoa(availablePort)
	}

	return bindings, nil
}

// CreateContainer 创建并启动容器 (修改后版本)
func CreateContainer(imageName string, containerName string, envVars []string, portBindings map[string]string) (string, error) {
	cli, err := getDockerClient()
	if err != nil {
		return "", err
	}

	// 构建容器配置
	containerConfig := &container.Config{
		Image: imageName,
		Env:   envVars,
	}

	// 构建主机配置
	hostConfig := &container.HostConfig{
		PortBindings: make(nat.PortMap),
	}

	// 添加端口映射
	for containerPort, hostPort := range portBindings {
		port, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			middleware.SugarLogger.Errorf("创建端口映射失败: %v", err)
			return "", err
		}
		hostConfig.PortBindings[port] = []nat.PortBinding{
			{
				HostPort: hostPort,
			},
		}
	}

	// 创建容器
	resp, err := cli.ContainerCreate(
		context.Background(),
		containerConfig,
		hostConfig,
		nil, // 网络配置
		nil, // 平台配置
		containerName,
	)
	if err != nil {
		middleware.SugarLogger.Errorf("创建容器失败: %v", err)
		return "", err
	}

	// 启动容器
	if err := cli.ContainerStart(context.Background(), resp.ID, container.StartOptions{}); err != nil {
		middleware.SugarLogger.Errorf("启动容器 %s 失败: %v", containerName, err)
		return "", err
	}

	// 保存容器信息
	containersMutex.Lock()
	createdContainers[resp.ID] = containerName
	containersMutex.Unlock()

	middleware.SugarLogger.Infof("成功创建并启动容器 %s (ID: %s)", containerName, resp.ID)
	return resp.ID, nil
}

// RemoveAllCreatedContainers 删除所有通过CreateContainer创建的容器
func RemoveAllCreatedContainers() error {
	containersMutex.Lock()
	defer containersMutex.Unlock()

	var lastError error
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	for containerID := range createdContainers {
		err := cli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{
			Force: true,
		})
		if err != nil {
			middleware.SugarLogger.Errorf("删除容器 %s 失败: %v", containerID, err)
			lastError = err
			continue
		}
		delete(createdContainers, containerID)
		middleware.SugarLogger.Infof("成功删除容器 %s", containerID)
	}

	return lastError
}

// RemoveContainer 删除容器
func RemoveContainer(containerID string, force bool) error {
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	options := container.RemoveOptions{
		Force: force, // 是否强制删除运行中的容器
	}

	if err := cli.ContainerRemove(context.Background(), containerID, options); err != nil {
		middleware.SugarLogger.Errorf("删除容器 %s 失败: %v", containerID, err)
		return err
	}

	middleware.SugarLogger.Infof("成功删除容器 %s (强制: %v)", containerID, force)
	return nil
}

// RemoveComposeContainers 删除通过Compose文件创建的所有容器和网络
func RemoveComposeContainers(composePath string) error {
	rawName := filepath.Base(filepath.Dir(composePath))
	stackName := normalizeProjectName(rawName)

	err := RemoveStackByName(stackName)
	if err != nil {
		return err
	}
	return nil
}

// 新增辅助函数：规范化项目名称
func normalizeProjectName(name string) string {
	// 1. 转换为小写
	name = strings.ToLower(name)

	// 2. 替换非法字符为下划线
	reg := regexp.MustCompile(`[^a-z0-9_-]`)
	name = reg.ReplaceAllString(name, "_")

	// 3. 确保不以非法字符开头
	if len(name) > 0 && !unicode.IsLetter(rune(name[0])) && !unicode.IsNumber(rune(name[0])) {
		name = "a" + name[1:]
	}

	// 4. 确保长度不超过64字符
	if len(name) > 64 {
		name = name[:64]
	}

	return name
}

// 检查服务是否已部署
// 检查服务是否已部署（只要存在容器就返回true，不检查运行状态）
func isServiceDeployed(_ *types.Project, serviceName string) bool {
	cli, err := getDockerClient()
	if err != nil {
		return false
	}

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: true, // 包含所有状态的容器
		Filters: filters.NewArgs(
			filters.Arg("label", fmt.Sprintf("com.docker.compose.service=%s", serviceName)),
		),
	})

	return err == nil && len(containers) > 0
}

// GetStackNames 获取Docker中所有存在的stackName
func GetStackNames() ([]string, error) {
	cli, err := getDockerClient()
	if err != nil {
		return nil, err
	}

	// 用于存储唯一的stackName
	stackNames := make(map[string]struct{})

	// 1. 从容器获取stackName
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("获取容器列表失败: %v", err)
	}

	for _, c := range containers {
		if project, ok := c.Labels["com.docker.compose.project"]; ok {
			stackNames[project] = struct{}{}
		}
	}

	// 2. 从网络获取stackName
	networks, err := cli.NetworkList(context.Background(), network.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("获取网络列表失败: %v", err)
	}

	for _, n := range networks {
		if project, ok := n.Labels["com.docker.compose.project"]; ok {
			stackNames[project] = struct{}{}
		}
	}

	// 转换为切片返回
	result := make([]string, 0, len(stackNames))
	for name := range stackNames {
		result = append(result, name)
	}

	return result, nil
}

// RemoveStackByName 根据stackName删除对应的容器、网络和卷
func RemoveStackByName(stackName string) error {
	cli, err := getDockerClient()
	if err != nil {
		return err
	}

	var lastError error

	// 1. 删除所有容器
	containers, err := cli.ContainerList(context.Background(), container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "com.docker.compose.project="+stackName),
		),
	})
	if err != nil {
		middleware.SugarLogger.Errorf("获取容器列表失败: %v", err)
		return err
	}

	for _, c := range containers {
		err = cli.ContainerRemove(context.Background(), c.ID, container.RemoveOptions{
			Force:         true,
			RemoveVolumes: true,
		})
		if err != nil {
			middleware.SugarLogger.Errorf("删除容器 %s 失败: %v", c.ID, err)
			lastError = err
		} else {
			middleware.SugarLogger.Infof("成功删除容器 %s", c.ID)
		}
	}

	// 2. 删除网络
	networks, err := cli.NetworkList(context.Background(), network.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "com.docker.compose.project="+stackName),
		),
	})
	if err != nil {
		middleware.SugarLogger.Errorf("获取网络列表失败: %v", err)
		return err
	}

	for _, n := range networks {
		err = cli.NetworkRemove(context.Background(), n.ID)
		if err != nil {
			middleware.SugarLogger.Errorf("删除网络 %s 失败: %v", n.ID, err)
			lastError = err
		} else {
			middleware.SugarLogger.Infof("成功删除网络 %s", n.ID)
		}
	}

	// 3. 删除卷(可选)
	volumes, err := cli.VolumeList(context.Background(), volume.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "com.docker.compose.project="+stackName),
		),
	})
	if err == nil { // 忽略卷查询错误
		for _, vol := range volumes.Volumes {
			err := cli.VolumeRemove(context.Background(), vol.Name, true)
			if err != nil {
				middleware.SugarLogger.Errorf("删除卷 %s 失败: %v", vol.Name, err)
				lastError = err
			} else {
				middleware.SugarLogger.Infof("成功删除卷 %s", vol.Name)
			}
		}
	}

	if lastError != nil {
		return fmt.Errorf("删除堆栈 %s 时发生部分错误: %v", stackName, lastError)
	}

	middleware.SugarLogger.Infof("成功删除堆栈 %s 的所有资源", stackName)
	return nil
}
