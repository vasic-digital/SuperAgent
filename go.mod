module dev.helix.agent

go 1.25.3

require github.com/gin-gonic/gin v1.12.0

require (
	dev.helix.agent/pkg/api v0.0.0-00010101000000-000000000000
	digital.vasic.agentic v0.0.0-00010101000000-000000000000
	digital.vasic.auth v0.0.0-00010101000000-000000000000
	digital.vasic.background v0.0.0
	digital.vasic.benchmark v0.0.0-00010101000000-000000000000
	digital.vasic.cache v0.0.0-00010101000000-000000000000
	digital.vasic.challenges v0.0.0
	digital.vasic.concurrency v0.0.0-00010101000000-000000000000
	digital.vasic.containers v0.0.0-00010101000000-000000000000
	digital.vasic.database v0.0.0-00010101000000-000000000000
	digital.vasic.debate v0.0.0-00010101000000-000000000000
	digital.vasic.docprocessor v0.0.0-00010101000000-000000000000
	digital.vasic.eventbus v0.0.0-00010101000000-000000000000
	digital.vasic.formatters v0.0.0-00010101000000-000000000000
	digital.vasic.helixmemory v0.0.0-00010101000000-000000000000
	digital.vasic.helixqa v0.0.0-00010101000000-000000000000
	digital.vasic.helixspecifier v0.0.0-00010101000000-000000000000
	digital.vasic.llmops v0.0.0-00010101000000-000000000000
	digital.vasic.llmorchestrator v0.0.0-00010101000000-000000000000
	digital.vasic.llmprovider v0.0.0
	digital.vasic.llmsverifier v0.0.0-00010101000000-000000000000
	digital.vasic.mcp v0.0.0-00010101000000-000000000000
	digital.vasic.memory v0.0.0-00010101000000-000000000000
	digital.vasic.messaging v0.0.0-00010101000000-000000000000
	digital.vasic.models v0.0.0
	digital.vasic.optimization v0.0.0-00010101000000-000000000000
	digital.vasic.planning v0.0.0-00010101000000-000000000000
	digital.vasic.plugins v0.0.0-00010101000000-000000000000
	digital.vasic.rag v0.0.0-00010101000000-000000000000
	digital.vasic.security v0.0.0-00010101000000-000000000000
	digital.vasic.selfimprove v0.0.0-00010101000000-000000000000
	digital.vasic.storage v0.0.0-00010101000000-000000000000
	digital.vasic.streaming v0.0.0-00010101000000-000000000000
	digital.vasic.toolschema v0.0.0-00010101000000-000000000000
	digital.vasic.vectordb v0.0.0-00010101000000-000000000000
	digital.vasic.visionengine v0.0.0-00010101000000-000000000000
	github.com/ClickHouse/ch-go v0.69.0
	github.com/ClickHouse/clickhouse-go/v2 v2.41.0
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/HelixDevelopment/HelixAgent/Toolkit v0.0.0-20260209162635-acd6d0755327
	github.com/alicebob/miniredis/v2 v2.36.1
	github.com/andybalholm/brotli v1.2.0
	github.com/beorn7/perks v1.0.1
	github.com/bytedance/gopkg v0.1.3
	github.com/bytedance/sonic v1.15.0
	github.com/bytedance/sonic/loader v0.5.0
	github.com/cenkalti/backoff/v5 v5.0.3
	github.com/cespare/xxhash/v2 v2.3.0
	github.com/cloudwego/base64x v0.1.6
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f
	github.com/dustin/go-humanize v1.0.1
	github.com/fsnotify/fsnotify v1.9.0
	github.com/gabriel-vasile/mimetype v1.4.12
	github.com/gin-contrib/sse v1.1.0
	github.com/go-faster/city v1.0.1
	github.com/go-faster/errors v0.7.1
	github.com/go-ini/ini v1.67.0
	github.com/go-logr/logr v1.4.3
	github.com/go-logr/stdr v1.2.2
	github.com/go-ole/go-ole v1.2.6
	github.com/go-playground/locales v0.14.1
	github.com/go-playground/universal-translator v0.18.1
	github.com/go-playground/validator/v10 v10.30.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/goccy/go-json v0.10.5
	github.com/goccy/go-yaml v1.19.2
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/gorilla/websocket v1.5.3
	github.com/graphql-go/graphql v0.8.1
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.7
	github.com/jackc/pgpassfile v1.0.0
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761
	github.com/jackc/pgx/v5 v5.7.6
	github.com/jackc/puddle/v2 v2.2.2
	github.com/joho/godotenv v1.5.1
	github.com/json-iterator/go v1.1.12
	github.com/klauspost/compress v1.18.3
	github.com/klauspost/cpuid/v2 v2.3.0
	github.com/klauspost/crc32 v1.3.0
	github.com/leodido/go-urn v1.4.0
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0
	github.com/mattn/go-isatty v0.0.20
	github.com/minio/crc64nvme v1.1.1
	github.com/minio/md5-simd v1.1.2
	github.com/minio/minio-go/v7 v7.0.98
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v1.0.2
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822
	github.com/ncruces/go-strftime v1.0.0
	github.com/neo4j/neo4j-go-driver/v5 v5.28.4
	github.com/paulmach/orb v0.12.0
	github.com/pelletier/go-toml/v2 v2.2.4
	github.com/philhofer/fwd v1.2.0
	github.com/pierrec/lz4/v4 v4.1.25
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c
	github.com/prometheus/client_golang v1.23.2
	github.com/prometheus/client_model v0.6.2
	github.com/prometheus/common v0.66.1
	github.com/prometheus/procfs v0.16.1
	github.com/quic-go/qpack v0.6.0
	github.com/quic-go/quic-go v0.59.0
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/redis/go-redis/v9 v9.17.2
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec
	github.com/rs/xid v1.6.0
	github.com/segmentio/asm v1.2.1
	github.com/segmentio/kafka-go v0.4.49
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/shoenig/go-m1cpu v0.1.6
	github.com/shopspring/decimal v1.4.0
	github.com/sirupsen/logrus v1.9.4
	github.com/stretchr/objx v0.5.2
	github.com/stretchr/testify v1.11.1
	github.com/tinylib/msgp v1.6.1
	github.com/tklauser/go-sysconf v0.3.12
	github.com/tklauser/numcpus v0.6.1
	github.com/twitchyliquid64/golang-asm v0.15.1
	github.com/ugorji/go/codec v1.3.1
	github.com/yuin/gopher-lua v1.1.1
	github.com/yusufpapurcu/wmi v1.2.4
	go.mongodb.org/mongo-driver/v2 v2.5.0
	go.opentelemetry.io/auto/sdk v1.2.1
	go.opentelemetry.io/otel v1.40.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.40.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.40.0
	go.opentelemetry.io/otel/exporters/stdout/stdouttrace v1.40.0
	go.opentelemetry.io/otel/metric v1.40.0
	go.opentelemetry.io/otel/sdk v1.40.0
	go.opentelemetry.io/otel/trace v1.40.0
	go.opentelemetry.io/proto/otlp v1.9.0
	go.uber.org/goleak v1.3.0
	go.uber.org/multierr v1.11.0
	go.uber.org/zap v1.27.1
	go.yaml.in/yaml/v2 v2.4.2
	go.yaml.in/yaml/v3 v3.0.4
	golang.org/x/arch v0.22.0
	golang.org/x/crypto v0.48.0
	golang.org/x/exp v0.0.0-20251023183803-a4bb9ffd2546
	golang.org/x/net v0.51.0
	golang.org/x/sync v0.20.0
	golang.org/x/sys v0.41.0
	golang.org/x/text v0.35.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260128011058-8636f8732409
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260128011058-8636f8732409
	google.golang.org/grpc v1.78.0
	google.golang.org/protobuf v1.36.11
	gopkg.in/yaml.v3 v3.0.1
	modernc.org/libc v1.67.6
	modernc.org/mathutil v1.7.1
	modernc.org/memory v1.11.0
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

replace digital.vasic.llmsverifier => ./LLMsVerifier/llm-verifier

replace digital.vasic.auth => ./Auth

replace digital.vasic.cache => ./Cache

replace digital.vasic.concurrency => ./Concurrency

replace digital.vasic.database => ./Database

replace digital.vasic.embeddings => ./Embeddings

replace digital.vasic.eventbus => ./EventBus

replace digital.vasic.helixmemory => ./HelixMemory

replace digital.vasic.helixspecifier => ./HelixSpecifier

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

replace digital.vasic.toolschema => ./ToolSchema

replace digital.vasic.skillregistry => ./SkillRegistry

replace digital.vasic.conversation => ./ConversationContext

replace digital.vasic.models => ./Models

replace digital.vasic.background => ./BackgroundTasks

replace digital.vasic.llmprovider => ./LLMProvider

replace digital.vasic.debate => ./DebateOrchestrator

replace digital.vasic.helixqa => ./HelixQA

replace digital.vasic.docprocessor => ./DocProcessor

replace digital.vasic.llmorchestrator => ./LLMOrchestrator

replace digital.vasic.visionengine => ./VisionEngine

replace github.com/HelixDevelopment/HelixAgent/Toolkit => ./Toolkit
