package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 定义了应用的全部配置项
// 使用 yaml 标签来匹配配置文件中的键
// 使用 env 标签来支持环境变量覆盖
type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Mode string `yaml:"mode"`
	} `yaml:"server"`
	JWT struct {
		Secret string `yaml:"secret"`
	} `yaml:"jwt"`
	WebSocket struct {
		ReadBufferSize  int    `yaml:"read_buffer_size"`
		WriteBufferSize int    `yaml:"write_buffer_size"`
		PingPeriod      string `yaml:"ping_period"`
		PongWait        string `yaml:"pong_wait"`
		MaxMessageSize  int    `yaml:"max_message_size"`
	} `yaml:"websocket"`
	Endpoints struct {
		PublicBaseURL string `yaml:"public_base_url"`
		WSPath        string `yaml:"ws_path"`
		AlarmPushPath string `yaml:"alarm_push_path"`
	} `yaml:"endpoints"`
}

// Load 从文件、环境变量加载配置，并应用默认值
func Load() (*Config, error) {
	// 1. 确定配置文件路径 (环境变量 > 环境 > 默认)
	configPath := getConfigPath()

	// 2. 从 YAML 文件读取基础配置
	cfg, err := loadFromFile(configPath)
	if err != nil {
		return nil, err
	}

	// 3. 应用默认值 (如果 YAML 中未定义)
	setDefaults(cfg)

	// 4. 应用环境变量覆盖 (最高优先级)
	applyEnvOverrides(cfg)

	// 5. 二次补齐默认值：处理 PORT 覆盖后 public_base_url 需要跟随端口的情况
	finalizeDefaults(cfg)

	// 6. 打印最终生效的配置
	log.Println("=== 施工现场告警通知中心 ===")
	log.Printf("配置来源: %s", configPath)
	log.Printf("服务模式: %s", cfg.Server.Mode)
	log.Printf("服务端口: %d", cfg.Server.Port)
	log.Printf("JWT 密钥: %s", maskString(cfg.JWT.Secret, 10))

	return cfg, nil
}


func getConfigPath() string {
	// 方案 A：优先读取 exe 同目录下的 config.yaml（发布/双击运行最稳定）
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		p := filepath.Join(dir, "config.yaml")
		if fileExists(p) {
			return p
		}
	}

	// 本地开发（go run）常见配置位置：项目根目录下的 config/config.yaml
	// 当 go run 时，os.Executable() 指向临时目录，这里用于回退避免找不到配置。
	if fileExists(filepath.Join("config", "config.yaml")) {
		return filepath.Join("config", "config.yaml")
	}

	// 最后兜底：当前工作目录下的 config.yaml
	return "config.yaml"
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

func loadFromFile(path string) (*Config, error) {
	var b []byte
	var err error

	b, err = os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件 %q 失败: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置 %q 失败: %w", path, err)
	}
	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "WanFang@JWT2024#SecureKey!ChangeInProduction"
	}
	if cfg.WebSocket.ReadBufferSize == 0 {
		cfg.WebSocket.ReadBufferSize = 1024
	}
	if cfg.WebSocket.WriteBufferSize == 0 {
		cfg.WebSocket.WriteBufferSize = 1024
	}
	if cfg.WebSocket.PingPeriod == "" {
		cfg.WebSocket.PingPeriod = "54s"
	}
	if cfg.WebSocket.PongWait == "" {
		cfg.WebSocket.PongWait = "60s"
	}
	if cfg.WebSocket.MaxMessageSize == 0 {
		cfg.WebSocket.MaxMessageSize = 512
	}

	if strings.TrimSpace(cfg.Endpoints.WSPath) == "" {
		cfg.Endpoints.WSPath = "/ws"
	}
	if strings.TrimSpace(cfg.Endpoints.AlarmPushPath) == "" {
		cfg.Endpoints.AlarmPushPath = "/api/push"
	}
}

func finalizeDefaults(cfg *Config) {
	if strings.TrimSpace(cfg.Endpoints.PublicBaseURL) == "" {
		cfg.Endpoints.PublicBaseURL = fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
	}
}

func applyEnvOverrides(cfg *Config) {
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, ok := parsePort(portStr); ok {
			cfg.Server.Port = p
		} else {
			log.Printf("[警告] 无效的环境变量 PORT: %q", portStr)
		}
	}
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWT.Secret = secret
	}

	if v := strings.TrimSpace(os.Getenv("PUBLIC_BASE_URL")); v != "" {
		cfg.Endpoints.PublicBaseURL = v
	}
	if v := strings.TrimSpace(os.Getenv("WS_PATH")); v != "" {
		cfg.Endpoints.WSPath = v
	}
	if v := strings.TrimSpace(os.Getenv("ALARM_PUSH_PATH")); v != "" {
		cfg.Endpoints.AlarmPushPath = v
	}
}

func parsePort(s string) (int, bool) {
	var p int
	_, err := fmt.Sscanf(s, "%d", &p)
	if err != nil || p <= 0 || p > 65535 {
		return 0, false
	}
	return p, true
}

func maskString(s string, showFirstN int) string {
	if len(s) == 0 {
		return "[未设置]"
	}
	if len(s) <= showFirstN {
		return "[已屏蔽]"
	}
	return s[:showFirstN] + "...[已屏蔽]"
}

