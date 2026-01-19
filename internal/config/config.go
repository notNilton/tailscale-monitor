package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config representa a configuração da aplicação
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Storage   StorageConfig   `yaml:"storage"`
	Metrics   MetricsConfig   `yaml:"metrics"`
	Tailscale TailscaleConfig `yaml:"tailscale"`
}

// ServerConfig configurações do servidor HTTP
type ServerConfig struct {
	Port          int  `yaml:"port"`
	TailscaleOnly bool `yaml:"tailscale_only"`
}

// StorageConfig configurações de armazenamento
type StorageConfig struct {
	Path          string `yaml:"path"`
	RetentionDays int    `yaml:"retention_days"`
}

// MetricsConfig configurações de coleta de métricas
type MetricsConfig struct {
	CollectionInterval time.Duration `yaml:"collection_interval"`
}

// TailscaleConfig configurações do Tailscale
type TailscaleConfig struct {
	APIKey  string `yaml:"api_key"`
	Tailnet string `yaml:"tailnet"`
	UseCLI  bool   `yaml:"use_cli"` // Se true, usa CLI; se false, usa API
}

// DefaultConfig retorna configuração padrão
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:          8080,
			TailscaleOnly: true,
		},
		Storage: StorageConfig{
			Path:          "/data/metrics.db",
			RetentionDays: 30,
		},
		Metrics: MetricsConfig{
			CollectionInterval: 30 * time.Second,
		},
		Tailscale: TailscaleConfig{
			APIKey:  os.Getenv("TAILSCALE_API_KEY"),
			Tailnet: os.Getenv("TAILSCALE_TAILNET"),
			UseCLI:  false, // Prefer API by default
		},
	}
}

// LoadConfig carrega configuração de arquivo YAML
func LoadConfig(path string) (*Config, error) {
	// Verifica se arquivo existe
	stat, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Arquivo não existe, usa config padrão
		return DefaultConfig(), nil
	}

	// Se existe mas é um diretório, usa config padrão
	if err == nil && stat.IsDir() {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// Se houver erro ao ler, usa config padrão
		return DefaultConfig(), nil
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}
