package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	mysql "github.com/go-sql-driver/mysql"

	"wklive/deploy/dbinit/models"
)

func main() {
	input, err := loadInput()
	must(err)

	db, err := sql.Open("mysql", loadMySqlDSN())
	must(err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	must(db.PingContext(ctx))

	result, err := models.NewDeliveryPreflightModel(db).Inspect(ctx, input)
	must(err)
	ready := printResult(result)
	if !ready {
		os.Exit(1)
	}
}

func loadInput() (models.DeliveryPreflightInput, error) {
	tenantID, err := parsePositiveInt64("DELIVERY_PREFLIGHT_TENANT_ID")
	if err != nil {
		return models.DeliveryPreflightInput{}, err
	}
	algorithm, err := parsePositiveInt64("DELIVERY_PREFLIGHT_FORMULA_ALGORITHM")
	if err != nil {
		return models.DeliveryPreflightInput{}, err
	}
	lookback, err := parsePositiveInt64("DELIVERY_PREFLIGHT_MAX_LOOKBACK_MS")
	if err != nil {
		return models.DeliveryPreflightInput{}, err
	}
	deviation, err := parsePositiveInt64("DELIVERY_PREFLIGHT_MAX_DEVIATION_BPS")
	if err != nil {
		return models.DeliveryPreflightInput{}, err
	}
	sourceCount, err := parsePositiveInt64("DELIVERY_PREFLIGHT_MIN_INPUT_COUNT")
	if err != nil {
		return models.DeliveryPreflightInput{}, err
	}
	interval, err := parsePositiveInt64("DELIVERY_PREFLIGHT_INTERVAL_MS")
	if err != nil {
		return models.DeliveryPreflightInput{}, err
	}
	return models.DeliveryPreflightInput{
		TenantID:               tenantID,
		Symbol:                 strings.TrimSpace(os.Getenv("DELIVERY_PREFLIGHT_SYMBOL")),
		SettlementAsset:        strings.TrimSpace(os.Getenv("DELIVERY_PREFLIGHT_SETTLEMENT_ASSET")),
		CategoryCode:           strings.TrimSpace(os.Getenv("DELIVERY_PREFLIGHT_CATEGORY_CODE")),
		Market:                 strings.TrimSpace(os.Getenv("DELIVERY_PREFLIGHT_MARKET")),
		PriceSymbol:            strings.TrimSpace(os.Getenv("DELIVERY_PREFLIGHT_PRICE_SYMBOL")),
		FormulaVersion:         strings.TrimSpace(os.Getenv("DELIVERY_PREFLIGHT_FORMULA_VERSION")),
		FormulaAlgorithm:       int(algorithm),
		FormulaMaxLookbackMs:   lookback,
		FormulaMaxDeviationBps: deviation,
		FormulaMinInputCount:   sourceCount,
		FormulaIntervalMs:      interval,
	}, nil
}

func printResult(result models.DeliveryPreflightResult) bool {
	ready := result.TechnicalReady()
	status := "FAIL"
	if ready {
		status = "PASS"
	}
	fmt.Printf("DELIVERY_TECHNICAL_PREFLIGHT=%s\n", status)
	fmt.Printf("DELIVERY_PRODUCTION_ENABLE_ALLOWED=false\n")
	fmt.Printf("DELIVERY_PRODUCT_COUNT=%d\n", result.ProductCount)
	fmt.Printf("DELIVERY_CONFIGURED_PRODUCT_COUNT=%d\n", result.ConfiguredProductCount)
	fmt.Printf("DELIVERY_SAFE_DISABLED_PRODUCT_COUNT=%d\n", result.SafeDisabledProductCount)
	fmt.Printf("DELIVERY_SYMBOL_ID=%d\n", result.SymbolID)
	fmt.Printf("DELIVERY_PRODUCT_STATUS=%d\n", result.ProductStatus)
	fmt.Printf("DELIVERY_SERVER_TIME_MS=%d\n", result.ServerTimeMs)
	fmt.Printf("DELIVERY_OPEN_CUTOFF_TIME_MS=%d\n", result.OpenCutoffTimeMs)
	fmt.Printf("DELIVERY_MATCHING_STOP_TIME_MS=%d\n", result.MatchingStopTimeMs)
	fmt.Printf("DELIVERY_TIME_MS=%d\n", result.DeliveryTimeMs)
	fmt.Printf("DELIVERY_ISOLATED_LEVERAGE_CONFIGS=%d\n", result.IsolatedLeverageConfigCount)
	fmt.Printf("DELIVERY_ISOLATED_LEVERAGE_DEFAULTS=%d\n", result.IsolatedLeverageDefaultCount)
	fmt.Printf("DELIVERY_CROSS_LEVERAGE_CONFIGS=%d\n", result.CrossLeverageConfigCount)
	fmt.Printf("DELIVERY_ENABLED_RISK_TIERS=%d\n", result.EnabledRiskTierCount)
	fmt.Printf("DELIVERY_VALID_RISK_TIERS=%d\n", result.ValidRiskTierCount)
	fmt.Printf("DELIVERY_BASE_RISK_TIERS=%d\n", result.BaseRiskTierCount)
	fmt.Printf("DELIVERY_RISK_COVERAGE_ENDS=%d\n", result.RiskCoverageEndCount)
	fmt.Printf("DELIVERY_FORMULAS=%d\n", result.FormulaCount)
	fmt.Printf("DELIVERY_CONFORMING_FORMULAS=%d\n", result.ConformingFormulaCount)
	fmt.Printf("DELIVERY_FRESH_SNAPSHOTS=%d\n", result.FreshSnapshotCount)
	fmt.Printf("DELIVERY_LATEST_SNAPSHOT_TIME_MS=%d\n", result.LatestSnapshotTimeMs)
	fmt.Printf("DELIVERY_ORDERS=%d\n", result.OrderCount)
	fmt.Printf("DELIVERY_FILLS=%d\n", result.FillCount)
	fmt.Printf("DELIVERY_POSITIONS=%d\n", result.PositionCount)
	fmt.Printf("DELIVERY_POSITION_HISTORY=%d\n", result.PositionHistoryCount)
	fmt.Printf("DELIVERY_RESERVATIONS=%d\n", result.ReservationCount)
	fmt.Printf("DELIVERY_SETTLEMENT_INSTRUCTIONS=%d\n", result.SettlementInstructionCount)
	fmt.Printf("DELIVERY_BATCHES=%d\n", result.DeliveryBatchCount)
	fmt.Printf("DELIVERY_SETTLEMENTS=%d\n", result.DeliverySettlementCount)
	fmt.Printf("DELIVERY_HISTORICAL_FACTS=%d\n", result.HistoricalFactCount())
	return ready
}

func loadMySqlDSN() string {
	if dsn := strings.TrimSpace(os.Getenv("MYSQL_DSN")); dsn != "" {
		return dsn
	}
	password := strings.TrimSpace(os.Getenv("MYSQL_PASSWORD"))
	if password == "" {
		password = getenv("MYSQL_ROOT_PASSWORD", "123456")
	}
	cfg := mysql.NewConfig()
	cfg.User = getenv("MYSQL_USER", "root")
	cfg.Passwd = password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(getenv("MYSQL_HOST", "mysql"), getenv("MYSQL_PORT", "3306"))
	cfg.DBName = getenv("MYSQL_DATABASE", "wklive")
	cfg.Params = map[string]string{"charset": "utf8mb4"}
	cfg.ParseTime = true
	cfg.Loc = time.Local
	return cfg.FormatDSN()
}

func parsePositiveInt64(name string) (int64, error) {
	value, err := strconv.ParseInt(strings.TrimSpace(os.Getenv(name)), 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return value, nil
}

func getenv(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
