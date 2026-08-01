package models

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

func TestFindAssignableShortsPageCapacityMySQL(t *testing.T) {
	dsn := os.Getenv("OPTION_ASSIGNMENT_PAGINATION_TEST_DSN")
	if dsn == "" {
		t.Skip("OPTION_ASSIGNMENT_PAGINATION_TEST_DSN is not set")
	}
	ctx := context.Background()
	conn := sqlx.NewMysql(dsn)
	const tenantBase int64 = 919060
	const contractID int64 = 919006

	cleanup := func() {
		_, _ = conn.ExecCtx(ctx, "DELETE FROM t_option_position WHERE tenant_id BETWEEN ? AND ?", tenantBase, tenantBase+3)
	}
	cleanup()
	t.Cleanup(cleanup)

	tests := []struct {
		name  string
		count int
	}{
		{name: "one", count: 1},
		{name: "exact_page", count: 500},
		{name: "page_plus_one", count: 501},
		{name: "five_thousand", count: 5000},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tenantID := tenantBase + int64(index)
			seedAssignablePositions(t, ctx, conn, tenantID, contractID, test.count)

			model := &customTOptionPositionModel{
				defaultTOptionPositionModel: &defaultTOptionPositionModel{
					CachedConn: sqlc.NewConnWithCache(conn, nil),
					table:      "`t_option_position`",
				},
			}
			cursorCreateTimes := int64(0)
			cursorID := int64(0)
			var items []*TOptionPosition
			for {
				page, err := model.FindAssignableShortsPage(
					ctx, tenantID, contractID, cursorCreateTimes, cursorID, 500,
				)
				if err != nil {
					t.Fatalf("query assignment page: %v", err)
				}
				if len(page) == 0 {
					break
				}
				for _, item := range page {
					if len(items) > 0 {
						previous := items[len(items)-1]
						if item.CreateTimes < previous.CreateTimes ||
							(item.CreateTimes == previous.CreateTimes && item.Id <= previous.Id) {
							t.Fatalf("non-FIFO page boundary: previous=%d/%d current=%d/%d",
								previous.CreateTimes, previous.Id, item.CreateTimes, item.Id)
						}
					}
					items = append(items, item)
				}
				cursorCreateTimes = page[len(page)-1].CreateTimes
				cursorID = page[len(page)-1].Id
				if len(page) < 500 {
					break
				}
			}
			if len(items) != test.count {
				t.Fatalf("assignable positions=%d want=%d", len(items), test.count)
			}
		})
	}
}

func seedAssignablePositions(
	t *testing.T,
	ctx context.Context,
	conn sqlx.SqlConn,
	tenantID, contractID int64,
	count int,
) {
	t.Helper()
	const batchSize = 500
	for start := 0; start < count; start += batchSize {
		end := start + batchSize
		if end > count {
			end = count
		}
		rows := make([]string, 0, end-start)
		args := make([]any, 0, (end-start)*10)
		for position := start; position < end; position++ {
			rows = append(rows, "(?,?,?,?,?,?,?,?,?,?)")
			createdAt := int64(1000 + position/2)
			args = append(args,
				tenantID, tenantID*10000+int64(position)+1, int64(1), contractID,
				int64(2), "1", "1", int64(1), createdAt, createdAt,
			)
		}
		query := `INSERT INTO t_option_position
(tenant_id,user_id,account_id,contract_id,side,position_qty,available_qty,status,create_times,update_times) VALUES ` +
			strings.Join(rows, ",")
		if _, err := conn.ExecCtx(ctx, query, args...); err != nil {
			t.Fatalf("seed positions %d-%d: %v", start, end, err)
		}
	}

	noise := []struct {
		suffix int64
		side   int64
		qty    string
		status int64
	}{
		{suffix: 1, side: 1, qty: "1", status: 1},
		{suffix: 2, side: 2, qty: "1", status: 2},
		{suffix: 3, side: 2, qty: "0", status: 1},
	}
	for _, item := range noise {
		if _, err := conn.ExecCtx(ctx, `INSERT INTO t_option_position
(tenant_id,user_id,account_id,contract_id,side,position_qty,available_qty,status,create_times,update_times)
VALUES(?,?,?,?,?,?,?, ?,9999,9999)`,
			tenantID, tenantID*10000+int64(count)+item.suffix, int64(1), contractID,
			item.side, item.qty, item.qty, item.status,
		); err != nil {
			t.Fatalf("seed non-assignable position: %v", err)
		}
	}
}
