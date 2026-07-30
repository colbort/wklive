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
	detailed := strings.EqualFold(os.Getenv("READINESS_DETAILED"), "true")
	input, err := loadInput(detailed)
	must(err)

	db, err := sql.Open("mysql", loadMySQLDSN())
	must(err)
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	must(db.PingContext(ctx))

	result, err := models.NewContractReadinessModel(db).Inspect(ctx, input, detailed)
	must(err)
	printResult(result)
}

func loadInput(detailed bool) (models.ContractReadinessInput, error) {
	if !detailed {
		return models.ContractReadinessInput{}, nil
	}
	tenantID, err := parsePositiveInt64("READINESS_TENANT_ID")
	if err != nil {
		return models.ContractReadinessInput{}, err
	}
	indexAlgorithm, err := parsePositiveInt64("READINESS_INDEX_ALGORITHM")
	if err != nil {
		return models.ContractReadinessInput{}, err
	}
	indexMaxDeviation, err := parsePositiveInt64("READINESS_INDEX_MAX_DEVIATION_BPS")
	if err != nil {
		return models.ContractReadinessInput{}, err
	}
	markMaxBasis, err := parsePositiveInt64("READINESS_MARK_MAX_BASIS_BPS")
	if err != nil {
		return models.ContractReadinessInput{}, err
	}
	formulaInterval, err := parsePositiveInt64("READINESS_FORMULA_INTERVAL_MS")
	if err != nil {
		return models.ContractReadinessInput{}, err
	}
	algorithm, err := parsePositiveInt64("READINESS_DELIVERY_ALGORITHM")
	if err != nil {
		return models.ContractReadinessInput{}, err
	}
	maxLookback, err := parsePositiveInt64("READINESS_MAX_LOOKBACK_MS")
	if err != nil {
		return models.ContractReadinessInput{}, err
	}
	maxDeviation, err := parsePositiveInt64("READINESS_MAX_DEVIATION_BPS")
	if err != nil {
		return models.ContractReadinessInput{}, err
	}
	sources := splitUnique(os.Getenv("READINESS_SOURCE_AUTHORITIES"))
	weights := splitValues(os.Getenv("READINESS_SOURCE_WEIGHTS"))
	return models.ContractReadinessInput{
		SourceAuthorities:         sources,
		SourceMarkets:             splitValues(os.Getenv("READINESS_SOURCE_MARKETS")),
		IndexSourceWeights:        splitValues(os.Getenv("READINESS_INDEX_SOURCE_WEIGHTS")),
		DeliverySourceWeights:     weights,
		ContractOncallAccount:     strings.TrimSpace(os.Getenv("READINESS_CONTRACT_ONCALL_ACCOUNT")),
		InsuranceOperatorAccount:  strings.TrimSpace(os.Getenv("READINESS_INSURANCE_OPERATOR_ACCOUNT")),
		DROperatorAccount:         strings.TrimSpace(os.Getenv("READINESS_DR_OPERATOR_ACCOUNT")),
		DeliveryOperatorAccount:   strings.TrimSpace(os.Getenv("READINESS_DELIVERY_OPERATOR_ACCOUNT")),
		ProductionReviewerAccount: strings.TrimSpace(os.Getenv("READINESS_PRODUCTION_REVIEWER_ACCOUNT")),
		ProductionApproverAccount: strings.TrimSpace(os.Getenv("READINESS_PRODUCTION_APPROVER_ACCOUNT")),
		CategoryCode:              strings.TrimSpace(os.Getenv("READINESS_CATEGORY_CODE")),
		Market:                    strings.TrimSpace(os.Getenv("READINESS_MARKET")),
		PriceSymbol:               strings.TrimSpace(os.Getenv("READINESS_PRICE_SYMBOL")),
		PerpetualSymbol:           strings.TrimSpace(os.Getenv("READINESS_PERPETUAL_SYMBOL")),
		DeliverySymbol:            strings.TrimSpace(os.Getenv("READINESS_DELIVERY_SYMBOL")),
		PerpetualPriceAuthority:   strings.TrimSpace(os.Getenv("READINESS_PERPETUAL_PRICE_AUTHORITY")),
		PerpetualPriceMarket:      strings.TrimSpace(os.Getenv("READINESS_PERPETUAL_PRICE_MARKET")),
		TenantID:                  tenantID,
		SettlementCoin:            strings.TrimSpace(os.Getenv("READINESS_SETTLEMENT_COIN")),
		InsuranceFundMinAvailable: strings.TrimSpace(os.Getenv("READINESS_INSURANCE_FUND_MIN_AVAILABLE")),
		IndexAlgorithm:            int(indexAlgorithm),
		IndexFormulaVersion:       strings.TrimSpace(os.Getenv("READINESS_INDEX_FORMULA_VERSION")),
		IndexMaxDeviationBps:      indexMaxDeviation,
		MarkFormulaVersion:        strings.TrimSpace(os.Getenv("READINESS_MARK_FORMULA_VERSION")),
		MarkMaxBasisBps:           markMaxBasis,
		MarkCurrentWeight:         strings.TrimSpace(os.Getenv("READINESS_MARK_CURRENT_WEIGHT")),
		MarkPreviousWeight:        strings.TrimSpace(os.Getenv("READINESS_MARK_PREVIOUS_WEIGHT")),
		FundingFormulaVersion:     strings.TrimSpace(os.Getenv("READINESS_FUNDING_FORMULA_VERSION")),
		PriceFormulaIntervalMs:    formulaInterval,
		DeliveryAlgorithm:         int(algorithm),
		DeliveryFormulaVersion:    strings.TrimSpace(os.Getenv("READINESS_FORMULA_VERSION")),
		DeliveryMaxLookbackMs:     maxLookback,
		DeliveryMaxDeviationBps:   maxDeviation,
	}, nil
}

func loadMySQLDSN() string {
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

func splitUnique(value string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func splitValues(value string) []string {
	result := make([]string, 0)
	for _, item := range strings.Split(value, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func printResult(result models.ContractReadinessResult) {
	fmt.Printf("READINESS_DB_DETAILED=%t\n", result.Detailed)
	fmt.Printf("READINESS_DB_ACTIVE_SOURCE_AUTHORITIES=%d\n", result.ActiveSourceAuthorityCount)
	fmt.Printf("READINESS_DB_DISTINCT_SOURCE_PROVIDERS=%d\n", result.DistinctSourceProviderCount)
	fmt.Printf("READINESS_DB_PUBLIC_REST_SOURCE_AUTHORITIES=%d\n", result.PublicRestSourceCount)
	fmt.Printf("READINESS_DB_PRICE_ENGINE_AUTHORITY=%d\n", result.PriceEngineAuthorityCount)
	fmt.Printf("READINESS_DB_INDEX_FORMULAS=%d\n", result.IndexFormulaCount)
	fmt.Printf("READINESS_DB_MARK_FORMULAS=%d\n", result.MarkFormulaCount)
	fmt.Printf("READINESS_DB_FUNDING_FORMULAS=%d\n", result.FundingFormulaCount)
	fmt.Printf("READINESS_DB_DELIVERY_FORMULAS=%d\n", result.DeliveryFormulaCount)
	fmt.Printf("READINESS_DB_FRESH_SOURCES=%d\n", result.FreshSourceCount)
	fmt.Printf("READINESS_DB_FRESH_OUTPUT_KINDS=%d\n", result.FreshOutputKindCount)
	fmt.Printf("READINESS_DB_INSURANCE_FUNDS=%d\n", result.InsuranceFundCount)
	fmt.Printf("READINESS_DB_FEE_REVENUE=%d\n", result.FeeRevenueCount)
	fmt.Printf("READINESS_DB_PERPETUAL_CONTRACTS=%d\n", result.PerpetualContractCount)
	fmt.Printf("READINESS_DB_DELIVERY_CONTRACTS=%d\n", result.DeliveryContractCount)
	fmt.Printf("READINESS_DB_INSURANCE_CONFIGS=%d\n", result.InsuranceConfigCount)
	fmt.Printf("READINESS_DB_CONTRACT_ONCALL_ACCOUNTS=%d\n", result.ContractOncallAccountCount)
	fmt.Printf("READINESS_DB_INSURANCE_OPERATORS=%d\n", result.InsuranceOperatorCount)
	fmt.Printf("READINESS_DB_DR_OPERATORS=%d\n", result.DROperatorCount)
	fmt.Printf("READINESS_DB_DELIVERY_OPERATORS=%d\n", result.DeliveryOperatorCount)
	fmt.Printf("READINESS_DB_PRODUCTION_REVIEWERS=%d\n", result.ProductionReviewerCount)
	fmt.Printf("READINESS_DB_PRODUCTION_APPROVERS=%d\n", result.ProductionApproverCount)
	fmt.Printf("READINESS_DB_PENDING_OUTBOX=%d\n", result.PendingOutboxCount)
	fmt.Printf("READINESS_DB_PROCESSING_OUTBOX=%d\n", result.ProcessingOutboxCount)
	fmt.Printf("READINESS_DB_FAILED_OUTBOX=%d\n", result.FailedOutboxCount)
	fmt.Printf("READINESS_DB_MANUAL_OUTBOX=%d\n", result.ManualOutboxCount)
	fmt.Printf("READINESS_DB_OPEN_OUTBOX=%d\n", result.OpenOutboxCount)
	fmt.Printf("READINESS_DB_UNHEALTHY_OUTBOX=%d\n", result.UnhealthyOutboxCount)
	fmt.Printf("READINESS_DB_OPEN_RECONCILIATION=%d\n", result.OpenReconciliationCount)
	fmt.Printf("READINESS_DB_OPEN_SETTLEMENT=%d\n", result.OpenSettlementCount)
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
