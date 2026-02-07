.PHONY: setup-py run-py run-go build-go

# --- Python Analytic Engine ---

# Setup: Bikin venv & install dependency otomatis
setup-py:
	cd analytic-engine-py && python3 -m venv venv && \
	. venv/bin/activate && \
	pip install --upgrade pip && \
	pip install -r requirements.txt
	@echo "✅ Python Setup Complete!"

# Run: Jalanin server Python pakai venv
run-py:
	cd analytic-engine-py && \
	. venv/bin/activate && \
	python main.py

# --- Golang Bot Service ---

# Setup Go (Download module)
setup-go:
	cd bot-service-go && go mod tidy
	@echo "✅ Go Setup Complete!"

# Run Bot
run-go:
	cd bot-service-go && go run cmd/bot/main.go
