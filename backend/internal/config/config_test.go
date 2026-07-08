package config

import "testing"

func TestValidate(t *testing.T) {
	cfg := Config{
		Server:  ServerConfig{Host: "0.0.0.0", Port: 8080, Mode: "debug"},
		MySQL:   MySQLConfig{Host: "127.0.0.1", Port: 3306, Username: "root", Password: "pwd", Database: "course_agent"},
		Redis:   RedisConfig{Addr: "127.0.0.1:6379"},
		Storage: StorageConfig{RootPath: "storage/course", QuotaBytes: 1 << 30},
		Agent:   AgentConfig{DefaultAgentName: "课程助教", PromptTemplate: "prompt"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}
