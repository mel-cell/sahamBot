
# Hapus semua cache & venv biar bersih total
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

run-py:
	cd analytic-engine-py && \
	. venv/bin/activate && \
	python main.py

setup-go:
	cd bot-service-go && go mod tidy
	@echo "✅ Go Setup Complete!"

run-go:
	cd bot-service-go && go run cmd/bot/main.go
