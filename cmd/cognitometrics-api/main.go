// file: cmd/server/main.go
package main

import (
	"CognitoMetrics/internal/api"
	"CognitoMetrics/internal/repository"
	"CognitoMetrics/internal/services"
	"log"
	"os" // 引入os包用于读取环境变量
)

func main() {
	log.Println("正在启动 CognitoMetrics 服务...")

	// 1. 初始化数据库仓库 (Repository)
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "cognitometrics.db" // 默认值
	}
	repo, err := repository.New(dbPath)
	if err != nil {
		log.Fatalf("无法初始化数据库仓库: %v", err)
	}
	log.Println("数据库连接和迁移成功。")

	// 2. 初始化所有服务 (Services)
	// 从环境变量加载配置，使应用更灵活、更安全
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	baseURL := os.Getenv("DEEPSEEK_BASE_URL")
	modelName := os.Getenv("DEEPSEEK_MODEL")
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	if modelName == "" {
		modelName = "deepseek-chat"
	}

	llmService, err := services.NewLLMService(apiKey, baseURL, modelName)
	if err != nil {
		log.Printf("警告：无法初始化LLM服务: %v。AI分析功能将不可用。", err)
	} else {
		log.Println("LLM服务初始化成功。")
	}

	aiAnalyzer := services.NewAIAnalyzer(repo, llmService)
	reportRunner := services.NewReportRunner(repo, aiAnalyzer)
	log.Println("核心服务层初始化完成。")

	// 3. 设置并注入路由 (Router & Handlers)
	router := api.SetupRouter(repo, reportRunner)
	log.Println("API路由设置完成。")

	// 4. 启动服务器
	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8000" // 默认端口
	}
	log.Printf("服务器启动，监听地址 http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
