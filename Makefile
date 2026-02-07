.PHONY: setup-py run-py run-go build-go clean-py run-dev

# --- Python Analytic Engine ---

# Setup: Bikin venv bersih & install dependency CPU-ONLY (Ringan)
clean-py:
	cd analytic-engine-py && rm -rf venv __pycache__ .pytest_cache
	rm -rf ~/.cache/pip
	@echo "🧹 Environment Python sudah bersih total!"

setup-py:
	@echo "🧹 Hapus environment lama..."
	make clean-py
	@echo "🏗️  Membuat environment baru..."
	cd analytic-engine-py && python3 -m venv venv && \
	. venv/bin/activate && \
	pip install --upgrade pip && \
	echo "⬇️  Menginstall PyTorch Versi CPU (Ringan)..." && \
	pip install torch --index-url https://download.pytorch.org/whl/cpu && \
	echo "📦 Menginstall sisa dependency..." && \
	pip install -r requirements.txt
	@echo "✅ Python Setup Complete (Versi Ringan)!"

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

# --- DEVELOPMENT (Run All) ---
run-dev:
	@echo "🚀 Starting Analytic Engine (Python) & Bot Service (Go)..."
	@trap 'kill 0' SIGINT; \
	make run-py & \
	sleep 5; \
	make run-go & \
	wait
