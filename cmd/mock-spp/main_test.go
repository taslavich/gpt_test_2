package loadtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gitlab.com/twinbid-exchange/RTB-exchange/internal/grpc/proto/types/ortb_V2_5"
)

// Конфигурация теста
const (
	sppAdapterURL = "http://localhost:8086/bid_v_2_5"
	threads       = 100              // Количество параллельных горутин
	targetRPS     = 10000            // Целевая нагрузка (RPS)
	testDuration  = 60 * time.Second // Длительность теста
)

var (
	globalIDCounter uint64
)

// Генератор тестовых BidRequest для ORTB 2.5
func generateBidRequest() *ortb_V2_5.BidRequest {
	bidFloor := float32(0.5)
	w := int32(300)
	h := int32(250)
	country := "US"

	id := atomic.AddUint64(&globalIDCounter, 1)

	return &ortb_V2_5.BidRequest{
		Id: stringPtr(fmt.Sprintf("req-%d", id)),
		At: int32Ptr(1),
		Imp: []*ortb_V2_5.Imp{
			{
				Id:       stringPtr("imp-1"),
				BidFloor: &bidFloor,
				Banner: &ortb_V2_5.Banner{
					W: &w,
					H: &h,
				},
			},
		},
		Device: &ortb_V2_5.Device{
			Ip: stringPtr(generateRandomIP()),
			Geo: &ortb_V2_5.Geo{
				Country: &country,
			},
		},
	}
}

// Вспомогательные функции
func stringPtr(s string) *string { return &s }
func int32Ptr(i int32) *int32    { return &i }

func generateRandomIP() string {
	// Простая генерация приватного IP, быстро меняющаяся
	n := atomic.AddUint64(&globalIDCounter, 1)
	return fmt.Sprintf("10.%d.%d.%d", byte(n>>16), byte(n>>8), byte(n))
}

// Тест производительности (rate-based)
func TestLoadRTBSystem(t *testing.T) {
	fmt.Printf("Starting load test: threads=%d targetRPS=%d duration=%v\n", threads, targetRPS, testDuration)

	// распределяем RPS по воркерам, учитывая остаток
	perWorker := targetRPS / threads
	remainder := targetRPS % threads

	// буфер результатов — targetRPS * duration (максимум). Берём минимум с разумным лимитом.
	maxResults := targetRPS * int(testDuration.Seconds())
	if maxResults < 1 {
		maxResults = targetRPS
	}
	if maxResults > 5_000_000 {
		maxResults = 5_000_000 // защита от OOM
	}

	results := make(chan *testResult, maxResults)
	var wg sync.WaitGroup

	startTime := time.Now()
	stopCh := make(chan struct{})

	// Запускаем воркеры
	for i := 0; i < threads; i++ {
		rps := perWorker
		if i < remainder {
			rps++
		}
		wg.Add(1)
		go workerRate(i, rps, results, &wg, stopCh)
	}

	// Останавливаем через duration
	time.AfterFunc(testDuration, func() {
		close(stopCh)
	})

	// Закрываем results после завершения воркеров
	go func() {
		wg.Wait()
		close(results)
	}()

	// Сбор результатов
	var resultsSlice []*testResult
	for result := range results {
		resultsSlice = append(resultsSlice, result)
	}

	totalTime := time.Since(startTime)
	analyzeResults(resultsSlice, totalTime)
}

// Воркер с таргетом rps
func workerRate(id, rps int, results chan<- *testResult, wg *sync.WaitGroup, stopCh <-chan struct{}) {
	defer wg.Done()
	if rps <= 0 {
		return
	}

	interval := time.Second / time.Duration(rps)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for {
		select {
		case <-stopCh:
			return
		case <-ticker.C:
			start := time.Now()
			bidRequest := generateBidRequest()
			result := sendBidRequestWithClient(bidRequest, start, client)
			// non-blocking send to avoid deadlock if results channel full
			select {
			case results <- result:
			default:
			}
		}
	}
}

// Отправка BidRequest с переиспользуемым клиентом
func sendBidRequestWithClient(bidRequest *ortb_V2_5.BidRequest, startTime time.Time, client *http.Client) *testResult {
	jsonData, err := json.Marshal(bidRequest)
	if err != nil {
		return &testResult{
			success:   false,
			latency:   time.Since(startTime),
			error:     err.Error(),
			timestamp: startTime,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", sppAdapterURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return &testResult{
			success:   false,
			latency:   time.Since(startTime),
			error:     err.Error(),
			timestamp: startTime,
		}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return &testResult{
			success:   false,
			latency:   time.Since(startTime),
			error:     err.Error(),
			timestamp: startTime,
		}
	}
	defer resp.Body.Close()

	latency := time.Since(startTime)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		return &testResult{
			success:   true,
			latency:   latency,
			error:     "",
			timestamp: startTime,
		}
	}

	return &testResult{
		success:   false,
		latency:   latency,
		error:     fmt.Sprintf("HTTP %d", resp.StatusCode),
		timestamp: startTime,
	}
}

// Результат теста
type testResult struct {
	success   bool
	latency   time.Duration
	error     string
	timestamp time.Time
}

// Анализ результатов
func analyzeResults(results []*testResult, totalTime time.Duration) {
	var totalLatency time.Duration
	var successCount int
	var errorCount int

	latencies := make([]time.Duration, 0, len(results))

	for _, result := range results {
		latencies = append(latencies, result.latency)
		totalLatency += result.latency

		if result.success {
			successCount++
		} else {
			errorCount++
			log.Printf("Error: %s (latency: %v)", result.error, result.latency)
		}
	}

	totalRequests := len(results)
	if totalRequests == 0 {
		fmt.Println("No requests recorded")
		return
	}

	rps := float64(totalRequests) / totalTime.Seconds()
	avgLatency := totalLatency / time.Duration(totalRequests)

	// Percentiles via sort.Slice
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p := func(q float64) time.Duration {
		if len(latencies) == 0 {
			return 0
		}
		idx := int(float64(len(latencies)) * q)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(latencies) {
			idx = len(latencies) - 1
		}
		return latencies[idx]
	}

	p50 := p(0.50)
	p95 := p(0.95)
	p99 := p(0.99)

	fmt.Printf("\n=== LOAD TEST RESULTS ===\n")
	fmt.Printf("Total Requests: %d\n", totalRequests)
	fmt.Printf("Successful: %d (%.2f%%)\n", successCount, float64(successCount)/float64(totalRequests)*100)
	fmt.Printf("Errors: %d (%.2f%%)\n", errorCount, float64(errorCount)/float64(totalRequests)*100)
	fmt.Printf("RPS: %.2f\n", rps)
	fmt.Printf("Test Duration: %v\n", totalTime)
	fmt.Printf("Average Latency: %v\n", avgLatency)
	fmt.Printf("P50 Latency: %v\n", p50)
	fmt.Printf("P95 Latency: %v\n", p95)
	fmt.Printf("P99 Latency: %v\n", p99)
	fmt.Printf("=========================\n")

	// Проверка требований 5k RPS — оставил для обратной совместимости
	if rps >= 5000 {
		fmt.Println("✅ Требование 5k RPS ВЫПОЛНЕНО")
	} else {
		fmt.Printf("❌ Требование 5k RPS НЕ ВЫПОЛНЕНО (текущее: %.2f)\n", rps)
	}

	if avgLatency < 100*time.Millisecond {
		fmt.Println("✅ Latency requirement MET")
	} else {
		fmt.Printf("⚠️  Latency is high: %v\n", avgLatency)
	}
}
