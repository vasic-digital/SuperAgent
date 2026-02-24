module dev.helix.agent

go 1.25.3

require github.com/gin-gonic/gin v1.11.0

require (
	dev.helix.agent/pkg/api v0.0.0-00010101000000-000000000000
	digital.vasic.agentic v0.0.0-00010101000000-000000000000
	digital.vasic.auth v0.0.0-00010101000000-000000000000
	digital.vasic.benchmark v0.0.0-00010101000000-000000000000
	digital.vasic.cache v0.0.0-00010101000000-000000000000
	digital.vasic.challenges v0.0.0-00010101000000-000000000000
	digital.vasic.concurrency v0.0.0-00010101000000-000000000000
	digital.vasic.containers v0.0.0-00010101000000-000000000000
	digital.vasic.database v0.0.0-00010101000000-000000000000
	digital.vasic.eventbus v0.0.0-00010101000000-000000000000
	digital.vasic.helixmemory v0.0.0-00010101000000-000000000000
	digital.vasic.formatters v0.0.0-00010101000000-000000000000
	digital.vasic.llmops v0.0.0-00010101000000-000000000000
	digital.vasic.mcp v0.0.0-00010101000000-000000000000
	digital.vasic.memory v0.0.0-00010101000000-000000000000
	digital.vasic.messaging v0.0.0-00010101000000-000000000000
	digital.vasic.optimization v0.0.0-00010101000000-000000000000
	digital.vasic.planning v0.0.0-00010101000000-000000000000
	digital.vasic.plugins v0.0.0-00010101000000-000000000000
	digital.vasic.rag v0.0.0-00010101000000-000000000000
	digital.vasic.security v0.0.0-00010101000000-000000000000
	digital.vasic.selfimprove v0.0.0-00010101000000-000000000000
	digital.vasic.storage v0.0.0-00010101000000-000000000000
	digital.vasic.streaming v0.0.0-00010101000000-000000000000
	digital.vasic.vectordb v0.0.0-00010101000000-000000000000
	github.com/ClickHouse/clickhouse-go/v2 v2.41.0
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/HelixDevelopment/HelixAgent/Toolkit v0.0.0-20260209162635-acd6d0755327
	github.com/alicebob/miniredis/v2 v2.36.1
	github.com/andybalholm/brotli v1.2.0
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/graphql-go/graphql v0.8.1
	github.com/jackc/pgx/v5 v5.7.6
	github.com/joho/godotenv v1.5.1
	github.com/minio/minio-go/v7 v7.0.98
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
	github.com/prometheus/client_golang v1.23.2
	github.com/quic-go/quic-go v0.57.1
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/redis/go-redis/v9 v9.17.2
	github.com/segmentio/kafka-go v0.4.49
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.10.2
	github.com/stretchr/testify v1.11.1
	go.opentelemetry.io/otel v1.40.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.40.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.40.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.40.0
	go.opentelemetry.io/otel/metric v1.40.0
	go.opentelemetry.io/otel/sdk v1.40.0
	go.opentelemetry.io/otel/trace v1.40.0
	go.uber.org/goleak v1.3.0
	go.uber.org/zap v1.27.1
	golang.org/x/crypto v0.48.0
	golang.org/x/sync v0.19.0
	golang.org/x/text v0.34.0
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
	gopkg.in/yaml.v3 v3.0.1
	llm-verifier v0.0.0-00010101000000-000000000000
	modernc.org/sqlite v1.44.2
)

replace dev.helix.agent/pkg/api => ./pkg/api

replace digital.vasic.containers => ./Containers

replace digital.vasic.challenges => ./Challenges

replace digital.vasic.agentic => ./Agentic

replace digital.vasic.llmops => ./LLMOps

replace digital.vasic.selfimprove => ./SelfImprove

replace digital.vasic.planning => ./Planning

replace digital.vasic.benchmark => ./Benchmark

replace llm-verifier => ./LLMsVerifier/llm-verifier

replace digital.vasic.auth => ./Auth

replace digital.vasic.cache => ./Cache

replace digital.vasic.concurrency => ./Concurrency

replace digital.vasic.database => ./Database

replace digital.vasic.embeddings => ./Embeddings

replace digital.vasic.eventbus => ./EventBus

replace digital.vasic.helixmemory => ./HelixMemory

replace digital.vasic.formatters => ./Formatters

replace digital.vasic.mcp => ./MCP_Module

replace digital.vasic.memory => ./Memory

replace digital.vasic.messaging => ./Messaging

replace digital.vasic.observability => ./Observability

replace digital.vasic.optimization => ./Optimization

replace digital.vasic.plugins => ./Plugins

replace digital.vasic.rag => ./RAG

replace digital.vasic.security => ./Security

replace digital.vasic.storage => ./Storage

replace digital.vasic.streaming => ./Streaming

replace digital.vasic.vectordb => ./VectorDB

require (
	github.com/ClickHouse/ch-go v0.69.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bytedance/sonic v1.14.0 // indirect
	github.com/bytedance/sonic/loader v0.3.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.6 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/gabriel-vasile/mimetype v1.4.8 // indirect
	github.com/gin-contrib/sse v1.1.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.27.0 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.7 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.18.3 // indirect
	github.com/klauspost/cpuid/v2 v2.3.0 // indirect
	github.com/klauspost/crc32 v1.3.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/minio/crc64nvme v1.1.1 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/ncruces/go-strftime v1.0.0 // indirect
	github.com/paulmach/orb v0.12.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/philhofer/fwd v1.2.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.25 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/tinylib/msgp v1.6.1 // indirect
	github.com/tklauser/go-sysconf v0.3.12 // indirect
	github.com/tklauser/numcpus v0.6.1 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.3.0 // indirect
	github.com/yuin/gopher-lua v1.1.1 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/proto/otlp v1.9.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/arch v0.20.0 // indirect
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546 // indirect
	golang.org/x/net v0.49.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20260128011058-8636f8732409 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260128011058-8636f8732409 // indirect
	modernc.org/libc v1.67.6 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)
