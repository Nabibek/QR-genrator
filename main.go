package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"QR-GENERATOR/internal/database"
	"QR-GENERATOR/internal/models"
	"QR-GENERATOR/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	qrcode "github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

func main() {
	// Загружаем .env файл
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  .env файл не найден, используются переменные окружения")
	}

	// Парси флагов командной строки
	seedFlag := flag.Bool("seed", false, "Заполнить БД тестовыми данными")
	genqrFlag := flag.Bool("genqr", false, "Сгенерировать QR коды")
	serverFlag := flag.Bool("server", true, "Запустить API сервер (по умолчанию)")
	flag.Parse()

	// Инициализируем подключение к БД
	if err := database.InitDB(); err != nil {
		log.Fatalf("❌ Ошибка подключения к БД: %v", err)
	}

	log.Println("✓ Подключено к БД warehouse")

	// Заполняем БД тестовыми данными если флаг --seed
	if *seedFlag {
		seedDatabase()
	}

	// Генерируем QR коды если флаг --genqr
	if *genqrFlag {
		generateTestQRCodes()
	}

	// Запускаем API сервер
	if *serverFlag || (!*seedFlag && !*genqrFlag) {
		startAPIServer()
	}
}

func seedDatabase() {
	db := database.GetDB()
	log.Println("🌱 Заполнение БД тестовыми данными...")

	// Создаём тестовые локации
	locations := []models.Location{
		{
			ID:          "location1",
			Code:        "LOC-A1",
			Description: "Полка A - Ряд 1",
			Row:         "A",
			Section:     "1",
			Shelf:       "1",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "location2",
			Code:        "LOC-A2",
			Description: "Полка A - Ряд 2",
			Row:         "A",
			Section:     "2",
			Shelf:       "1",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "location3",
			Code:        "LOC-B1",
			Description: "Полка B - Ряд 1",
			Row:         "B",
			Section:     "1",
			Shelf:       "2",
			CreatedAt:   time.Now(),
		},
	}

	for _, loc := range locations {
		result := db.FirstOrCreate(&loc, models.Location{ID: loc.ID})
		if result.Error != nil {
			log.Printf("❌ Ошибка при создании локации %s: %v", loc.ID, result.Error)
		}
	}
	log.Printf("✓ Создано %d локаций", len(locations))

	// Создаём тестовые товары
	items := []models.Item{
		{
			ID:          "item1",
			Name:        "Widget Pro",
			SKU:         "WDGT-001",
			Description: "Высокопроизводительный виджет",
			Quantity:    50,
			PartNumber:  "PN-2024-001",
			BatchNumber: "BATCH-2024-01",
			LocationID:  "location1",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "item2",
			Name:        "Gadget Plus",
			SKU:         "GDGT-002",
			Description: "Улучшенный гаджет с функциями",
			Quantity:    30,
			PartNumber:  "PN-2024-002",
			BatchNumber: "BATCH-2024-01",
			LocationID:  "location2",
			CreatedAt:   time.Now(),
		},
		{
			ID:          "item3",
			Name:        "Component X",
			SKU:         "COMP-003",
			Description: "Существенный компонент",
			Quantity:    100,
			PartNumber:  "PN-2024-003",
			BatchNumber: "BATCH-2024-02",
			LocationID:  "location3",
			CreatedAt:   time.Now(),
		},
	}

	for _, item := range items {
		result := db.FirstOrCreate(&item, models.Item{ID: item.ID})
		if result.Error != nil {
			log.Printf("❌ Ошибка при создании товара %s: %v", item.ID, result.Error)
		}
	}
	log.Printf("✓ Создано %d товаров", len(items))

	// Создаём тестового пользователя
	user := models.User{
		ID:           "user1",
		Username:     "operator1",
		Email:        "operator1@warehouse.local",
		PasswordHash: "sha_hash_of_password123",
		Role:         "operator",
		CreatedAt:    time.Now(),
	}

	result := db.FirstOrCreate(&user, models.User{ID: user.ID})
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		log.Printf("❌ Ошибка при создании пользователя: %v", result.Error)
	} else {
		log.Println("✓ Создан тестовый пользователь (оператор)")
	}
}

func generateTestQRCodes() {
	// Создаём папку для QR кодов
	qrcodeDir := "qrcodes"
	if err := os.MkdirAll(qrcodeDir, 0755); err != nil {
		log.Fatalf("❌ Ошибка создания папки qrcodes: %v", err)
	}

	log.Println("📱 Генерирование QR кодов с простым формат ID...")

	// Генерируем QR коды для товаров (формат: ITEM:item123)
	for i := 1; i <= 3; i++ {
		itemID := fmt.Sprintf("item%d", i)
		content := fmt.Sprintf("ITEM:%s", itemID)
		filePath := filepath.Join(qrcodeDir, fmt.Sprintf("item_%s.png", itemID))

		err := qrcode.WriteFile(content, qrcode.High, 256, filePath)
		if err != nil {
			log.Printf("❌ Ошибка генерирования QR для товара %s: %v", itemID, err)
		} else {
			log.Printf("✓ Сгенерирован QR: %s (содержание: %s)", filePath, content)
		}
	}

	// Генерируем QR коды для локаций (формат: LOC:location123)
	for i := 1; i <= 3; i++ {
		locID := fmt.Sprintf("location%d", i)
		content := fmt.Sprintf("LOC:%s", locID)
		filePath := filepath.Join(qrcodeDir, fmt.Sprintf("location_%s.png", locID))

		err := qrcode.WriteFile(content, qrcode.High, 256, filePath)
		if err != nil {
			log.Printf("❌ Ошибка генерирования QR для локации %s: %v", locID, err)
		} else {
			log.Printf("✓ Сгенерирован QR: %s (содержание: %s)", filePath, content)
		}
	}

	log.Println("✓ Все QR коды сгенерированы в папке ./qrcodes/")
}

func startAPIServer() {
	// Устанавливаем режим Gin (release для production, debug для development)
	serverEnv := os.Getenv("SERVER_ENV")
	if serverEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Создаём Gin роутер
	router := gin.Default()

	// Добавляем CORS middleware (нужен для браузера)
	router.Use(corsMiddleware())

	// Регистрируем все маршруты
	routes.SetupRoutes(router)

	// Получаем порт из переменных окружения
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	serverAddr := fmt.Sprintf(":%s", port)

	log.Printf("\n🚀 API сервер запущен на http://localhost:%s", port)
	log.Printf("\n📱 Сканер доступен: http://localhost:%s", port)
	log.Printf("\n📚 Документация API:", port)
	log.Printf("   POST   /api/login         - Вход (username/password)")
	log.Printf("   GET    /api/item/:id      - Получить информацию о товаре")
	log.Printf("   GET    /api/item/:id/history - История перемещений товара")
	log.Printf("   POST   /api/move          - Переместить товар на новую локацию")
	log.Printf("   GET    /health            - Проверка статуса сервера\n")

	// Запускаем сервер
	if err := router.Run(serverAddr); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Ошибка запуска сервера: %v", err)
	}
}

// corsMiddleware добавляет CORS заголовки для браузера
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
