package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	tele "gopkg.in/telebot.v3"
	
	"github.com/saham-analytic-bot/bot/internal/repository/db"
)

// --- Structs for Analyic Engine Response ---
type AnalysisResult struct {
	Symbol       string              `json:"symbol"`
	CurrentPrice float64             `json:"current_price"`
	Technical    TechnicalIndicators `json:"technical"`
	Summary      string              `json:"summary"`
}

type TechnicalIndicators struct {
	RSI        float64 `json:"rsi"`
	MACD       float64 `json:"macd"`
	MACDSignal float64 `json:"macd_signal"`
	Signal     string  `json:"signal"`
}

// Global Variables
var (
	bot         *tele.Bot
	analyticURL string
	repo        *db.Repository // Database Access
)

func main() {
	// 1. Load Environment Variables
	if err := godotenv.Load(".env"); err != nil {
		if err := godotenv.Load("../.env"); err != nil {
			log.Println("Warning: Error loading .env file, checking system env")
		}
	}

	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	analyticURL = os.Getenv("ANALYTIC_ENGINE_URL")

	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}
	if analyticURL == "" {
		analyticURL = "http://localhost:8000"
	}

	// 2. Initialize Database (SQLite)
	var err error
	repo, err = db.NewRepository("./saham.db")
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	log.Println("✅ Database Connected (saham.db)")

	// 3. Setup Bot Settings
	pref := tele.Settings{
		Token:  botToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err = tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Println("🤖 Bot Started! Listening for messages...")

	// 4. Setup Cron Scheduler (Auto Check)
	c := cron.New()
	
	// Jalankan setiap jam (example: "0 * * * *") -> Tiap menit ke-0
	// Untuk demo, kita set tiap 30 menit: "*/30 * * * *"
	c.AddFunc("*/30 * * * *", runSchedulerTask) 
	c.Start()
	log.Println("⏰ Scheduler Active (Every 30 mins)")

	// --- HANDLERS ---
	
	// /start
	bot.Handle("/start", func(c tele.Context) error {
		return c.Send("Halo! Saya Bot Saham Pintar 🧠.\n\n" +
			"**Perintah:**\n" +
			"`/analisa [KODE]` - Cek saham sekarang\n" +
			"`/pantau [KODE]` - Masukkan ke Watchlist\n" +
			"`/hapus [KODE]` - Hapus dari Watchlist\n" +
			"`/list` - Lihat Watchlist kamu")
	})

	// /analisa [SYMBOL]
	bot.Handle("/analisa", func(c tele.Context) error {
		symbol := extractSymbol(c)
		if symbol == "" {
			return c.Send("Mohon masukkan kode saham.\nContoh: `/analisa BBRI.JK`")
		}
        
        msg, err := bot.Send(c.Chat(), fmt.Sprintf("⏳ Sedang menganalisa %s...", symbol))
        // ... (Error handling omitted for brevity)

		result, err := fetchAnalysis(analyticURL, symbol)
		if err != nil {
			if msg != nil {
            	bot.Edit(msg, fmt.Sprintf("❌ Gagal menganalisa %s.\nError: %v", symbol, err))
			}
			return err
		}

		reply := formatAnalysisMessage(result)
		if msg != nil {
			_, err := bot.Edit(msg, reply, tele.ModeMarkdown)
			return err
		}
		return c.Send(reply, tele.ModeMarkdown)
	})

	// /pantau [SYMBOL] (Add to Watchlist)
	bot.Handle("/pantau", func(c tele.Context) error {
		symbol := extractSymbol(c)
		if symbol == "" {
			return c.Send("❌ Masukkan kode saham.\nContoh: `/pantau BBCA.JK`")
		}

		// Insert to DB
		err := repo.AddWatch(c.Sender().ID, symbol)
		if err != nil {
			return c.Send(fmt.Sprintf("❌ Gagal menyimpan %s (Mungkin sudah ada?).", symbol))
		}
		
		return c.Send(fmt.Sprintf("✅ **%s** berhasil ditambahkan ke Watchlist!\nSaya akan lapor kalau ada sinyal BUY/SELL.", symbol), tele.ModeMarkdown)
	})

	// /hapus [SYMBOL] (Remove from Watchlist)
	bot.Handle("/hapus", func(c tele.Context) error {
		symbol := extractSymbol(c)
		if symbol == "" {
			return c.Send("❌ Masukkan kode saham.\nContoh: `/hapus BBCA.JK`")
		}

		// Delete from DB
		err := repo.RemoveWatch(c.Sender().ID, symbol)
		if err != nil {
			return c.Send("❌ Gagal menghapus data.")
		}
		
		return c.Send(fmt.Sprintf("🗑️ **%s** dihapus dari Watchlist.", symbol), tele.ModeMarkdown)
	})

	// /list (Show Watchlist)
	bot.Handle("/list", func(c tele.Context) error {
		symbols, err := repo.GetUserWatchlist(c.Sender().ID)
		if err != nil {
			return c.Send("❌ Gagal mengambil data watchlist.")
		}

		if len(symbols) == 0 {
			return c.Send("📭 Watchlist kamu masih kosong.\nPakai `/pantau [KODE]` untuk tambah.")
		}

		msg := "📋 **Watchlist Saham Kamu:**\n"
		for i, s := range symbols {
			msg += fmt.Sprintf("%d. `%s`\n", i+1, s)
		}
		return c.Send(msg, tele.ModeMarkdown)
	})

	// 5. Start Bot
	bot.Start()
}


// --- HELPER FUNCTIONS ---

func extractSymbol(c tele.Context) string {
	args := c.Args()
	if len(args) > 0 {
		return strings.ToUpper(args[0])
	}
	return ""
}

func formatAnalysisMessage(result *AnalysisResult) string {
	signalIcon := "⚪"
	if strings.Contains(result.Technical.Signal, "BUY") {
		signalIcon = "🟢"
	} else if strings.Contains(result.Technical.Signal, "SELL") {
		signalIcon = "🔴"
	}

	return fmt.Sprintf(
		"📊 **Analisa Saham: %s**\n\n"+
			"💰 **Harga:** Rp %.0f\n"+
			"📈 **RSI:** %.2f\n"+
			"🌊 **MACD:** %.2f (Signal: %.2f)\n\n"+
			"%s **Sinyal AI:** %s\n"+
			"📝 **Ringkasan:**\n%s",
		result.Symbol,
		result.CurrentPrice,
		result.Technical.RSI,
		result.Technical.MACD,
		result.Technical.MACDSignal,
		signalIcon,
		result.Technical.Signal,
		result.Summary,
	)
}

// Function to call Python Analytic Engine
func fetchAnalysis(baseURL, symbol string) (*AnalysisResult, error) {
	url := fmt.Sprintf("%s/analyze/%s", baseURL, symbol)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("Connection error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API Error: %s", resp.Status)
	}

	var result AnalysisResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("Decode error: %v", err)
	}

	return &result, nil
}

// --- SCHEDULER LOGIC ---

func runSchedulerTask() {
	log.Println("⏰ Running Scheduled Watchlist Check...")
	
	// 1. Get All Unique Users & Symbols
	// (Implementasi sederhana: Loop semua user -> Loop semua symbol)
	// Idealnya batch processing, tapi ini cukup untuk MVP.
	
	// Ambil semua data user->symbols (kita asumsikan fungsi repository mendukung ini)
	// Tapi repository kita tadi belum punya GetALL.
	// Oh, di step sebelumnya kita udah bikin GetAllUniqueUserSymbols() ! Mantap.
	
	userMap, err := repo.GetAllUniqueUserSymbols()
	if err != nil {
		log.Printf("Scheduler Error: %v", err)
		return
	}

	for userID, symbols := range userMap {
		for _, symbol := range symbols {
			// Analisa Saham
			result, err := fetchAnalysis(analyticURL, symbol)
			if err != nil {
				log.Printf("Failed to analyze %s for user %d: %v", symbol, userID, err)
				continue
			}

			// LOGIC PENTING: Kapan harus notif?
			// Kita notif HANYA jika sinyal KUAT (Strong Buy / Strong Sell)
			// Atau RSI Oversold/Overbought.
			shouldNotify := false
			
			if strings.Contains(result.Technical.Signal, "STRONG") {
				shouldNotify = true
			}
			if result.Technical.RSI < 30 || result.Technical.RSI > 70 {
				shouldNotify = true
			}

			if shouldNotify {
				// Kirim Pesan ke User
				msg := fmt.Sprintf("🚨 **ALERT: %s**\n\nSinyal: %s\nRSI: %.2f\n\nCek detail: `/analisa %s`", 
					symbol, result.Technical.Signal, result.Technical.RSI, symbol)
				
				// Send direct message (User ID must be Chat ID)
				// Telebot butuh Recipient interface. Kita bungkus int64 jadi Recipient.
				recipient := &tele.User{ID: userID}
				bot.Send(recipient, msg, tele.ModeMarkdown)
				
				log.Printf("🔔 Sent alert to %d for %s", userID, symbol)
				
				// Sleep biar gak spam server
				time.Sleep(2 * time.Second)
			}
		}
	}
}
