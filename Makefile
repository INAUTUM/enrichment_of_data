# Имя приложения
APP_NAME=myapp

# Путь к файлу
MAIN_FILE=main.go

# Папка для сборки (если нужно)
BUILD_DIR=build

# Настройки среды
GOOS=darwin
GOARCH=amd64

# Команды
.PHONY: build run test clean install

# Сборка приложения
build:
	@echo "==> Building $(APP_NAME)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)
	@echo "==> Build complete."

# Запуск приложения
run_no:
	@echo "==> Running $(APP_NAME)..."
	# go run $(MAIN_FILE) < input.txt
	go run $(MAIN_FILE)

run:
	@echo "==> Running $(APP_NAME)..."
	go run $(MAIN_FILE) < input.txt

run_db:
    export $(cat .env | xargs);go run main.go

# Запуск тестов
test:
	@echo "==> Running tests..."
	go test ./... -v

# Очистка сборки
clean:
	@echo "==> Cleaning up..."
	rm -rf $(BUILD_DIR)
	@echo "==> Clean complete."

# Установка приложения
install:
	@echo "==> Installing $(APP_NAME)..."
	go install $(MAIN_FILE)
	@echo "==> Install complete."