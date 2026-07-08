package config

import "testing"

func TestValidate(t *testing.T) {
	cfg := Config{
		Server: ServerConfig{Host: "0.0.0.0", Port: 8080, Mode: "debug"},
		MySQL:  MySQLConfig{Host: "127.0.0.1", Port: 3306, Username: "root", Database: "course_agent"},
		Redis:  RedisConfig{Addr: "127.0.0.1:6379"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got error: %v", err)
	}
}
