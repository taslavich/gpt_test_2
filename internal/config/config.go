package config

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

// Кастомный тип для map[string]string
type MapStringToString map[string]string

func (m *MapStringToString) SetValue(value string) error {
	*m = make(MapStringToString)
	if value == "" {
		return nil
	}

	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			(*m)[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return nil
}

// Кастомный тип для []string
type ListString []string

func (l *ListString) SetValue(value string) error {
	*l = make(ListString, 0)
	if value == "" {
		return nil
	}

	items := strings.Split(value, ",")
	for _, item := range items {
		*l = append(*l, strings.TrimSpace(item))
	}
	return nil
}

type BiddingEngineConfig struct {
	HttpServer
	ProfitPercent  float32 `yaml:"PROFIT_PERCENT" env:"PROFIT_PERCENT" env-default:"0.2"`
	SystemHostname string  `yaml:"SYSTEM_HOSTNAME" env:"SYSTEM_HOSTNAME"`
	RedisConfig
}

type RouterConfig struct {
	HttpServer
	DSPEndpoints_v_2_4 ListString `yaml:"DSP_ENDPOINTS_V_2_4" env:"DSP_ENDPOINTS_V_2_4"`
	//SPPEndpoints_v_2_4 ListString        `yaml:"SPP_ENDPOINTS_V_2_4" env:"SPP_ENDPOINTS_V_2_4"`
	DspRulesConfigPath string `yaml:"DSP_RULES_CONFIG_PATH" env:"DSP_RULES_CONFIG_PATH"`
	SppRulesConfigPath string `yaml:"SPP_RULES_CONFIG_PATH" env:"SPP_RULES_CONFIG_PATH"`

	BidResponsesTimeout time.Duration `yaml:"BID_RESPONSES_TIMEOUT" env:"BID_RESPONSES_TIMEOUT"`

	RedisConfig
}

type OrchestratorConfig struct {
	HttpServer
	UriOfBidEngine string        `yaml:"URI_OF_BID_ENGINE" env:"URI_OF_BID_ENGINE"`
	UriOfDspRouter string        `yaml:"URI_OF_DSP_ROUTER" env:"URI_OF_DSP_ROUTER"`
	AuctionTimeout time.Duration `yaml:"AUCTION_TIMEOUT" env:"AUCTION_TIMEOUT"`
	GetBidsTimeout time.Duration `yaml:"GET_BIDS_TIMEOUT" env:"GET_BIDS_TIMEOUT"`

	RedisConfig
}

type SppAdapterConfig struct {
	HttpServer
	UriOfOrchestrator   string        `yaml:"URI_OF_ORCHESTRATOR" env:"URI_OF_ORCHESTRATOR"`
	NurlTimeout         time.Duration `yaml:"NURL_TIMEOUT" env:"NURL_TIMEOUT"`
	BurlTimeout         time.Duration `yaml:"BURL_TIMEOUT" env:"BURL_TIMEOUT"`
	GetWinnerBidTimeout time.Duration `yaml:"GET_WINNER_BID_TIMEOUT" env:"GET_WINNER_BID_TIMEOUT"`
	GeoIpDbPath         string        `yaml:"GEO_IP_DB_PATH" env:"GEO_IP_DB_PATH"`

	RedisConfig
}

type KafkaLoaderConfig struct {
	RedisConfig
	KafkaConfig
	BatchSize int64 `yaml:"BATCH_SIZE" env:"BATCH_SIZE"`
}

type ClickhouseConfig struct {
	ClickHouseTable string `yaml:"CLICK_HOUSE_TABLE" env:"CLICK_HOUSE_TABLE"`
	BatchSize       int    `yaml:"CLICKHOUSE_BATCH_SIZE" env:"CLICKHOUSE_BATCH_SIZE"`
	Username        string `yaml:"CLICKHOUSE_USERNAME" env:"CLICKHOUSE_USERNAME"`
	Password        string `yaml:"CLICKHOUSE_PASSWORD" env:"CLICKHOUSE_PASSWORD"`
	Host            string `yaml:"CLICKHOUSE_HOST" env:"CLICKHOUSE_HOST" env-default:"hntzp0jsnf.europe-west4.gcp.clickhouse.cloud"`
	Port            string `yaml:"CLICKHOUSE_PORT" env:"CLICKHOUSE_PORT" env-default:"9440"`
	Database        string `yaml:"CLICKHOUSE_DB" env:"CLICKHOUSE_DB" env-default:"rtb"`
}

type ClickhouseLoaderConfig struct {
	Kafka      KafkaConfig
	Clickhouse ClickhouseConfig
	TimeoutSec int `yaml:"TIMEOUT_SEC" env:"TIMEOUT_SEC"`
}

type MockDspConfig struct {
	HttpServer
	DspName string  `env:"DSP_NAME"`
	Price   float32 `env:"PRICE"`
	Adid    string  `env:"ADID"`
}

type RedisConfig struct {
	RedisHost     string `yaml:"REDIS_HOST" env:"REDIS_HOST"`
	RedisPort     string `yaml:"REDIS_PORT" env:"REDIS_PORT"`
	RedisDB       int    `yaml:"REDIS_DB" env:"REDIS_DB"`
	RedisPassword string `yaml:"REDIS_PASSWORD" env:"REDIS_PASSWORD"`
}

type KafkaConfig struct {
	KafkaBroker      string `yaml:"KAFKA_BROKER" env:"KAFKA_BROKER"`
	KafkaTopic       string `yaml:"KAFKA_TOPIC" env:"KAFKA_TOPIC"`
	FlushIntervalSec int    `yaml:"FLUSH_INTERVAL_SEC" env:"FLUSH_INTERVAL_SEC"`
	KafkaGroupID     string `yaml:"KAFKA_GROUP_ID" env:"KAFKA_GROUP_ID"`
}

type HttpServer struct {
	Host string `yaml:"HOSTNAME" env:"HOSTNAME" env-default:"127.0.0.1"`
	Port uint16 `yaml:"PORT" env:"PORT" env-default:"8080"`
}

func getEnvFileNames() []string {
	return []string{".env.local", ".env"}
}

func LoadConfig[
	T BiddingEngineConfig |
		RouterConfig |
		SppAdapterConfig |
		OrchestratorConfig |
		KafkaLoaderConfig |
		ClickhouseLoaderConfig |
		MockDspConfig,
](ctx context.Context) (*T, error) {
	for _, fileName := range getEnvFileNames() {
		err := godotenv.Load(fileName)
		if err != nil {
			log.Printf("error loading %s fileName : %v", fileName, err)
		}
	}

	var cfg T
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
