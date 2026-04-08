package models

import (
    "context"
    "fmt"
)

type TradeSymbolModel interface {
    tTradeSymbolModel
    FindPage(ctx context.Context, cursor int64, limit int64) ([]*TTradeSymbol, int64, error)
}

func (m *defaultTTradeSymbolModel) FindPage(ctx context.Context, cursor int64, limit int64) ([]*TTradeSymbol, int64, error) {
    if limit <= 0 {
        limit = 10
    }
    if limit > 100 {
        limit = 100
    }

    where := "1=1"
    args := make([]any, 0, 2)

    // ---- total ----
    var total int64
    countSql := fmt.Sprintf("SELECT COUNT(1) FROM %s WHERE %s", m.table, where)
    if err := m.QueryRowNoCacheCtx(ctx, &total, countSql, args...); err != nil {
        return nil, 0, err
    }

    listArgs := append([]any{}, args...)
    var listSql string

    if cursor <= 0 {
        listSql = fmt.Sprintf(
            `SELECT %s
            FROM %s
            WHERE %s
            ORDER BY id DESC
            LIMIT ?`,
            tTradeSymbolRows, m.table, where,
        )
        listArgs = append(listArgs, limit)
    } else {
        listSql = fmt.Sprintf(
            `SELECT %s
            FROM %s
            WHERE %s AND id < ?
            ORDER BY id DESC
            LIMIT ?`,
            tTradeSymbolRows, m.table, where,
        )
        listArgs = append(listArgs, cursor, limit)
    }

    var list []*TTradeSymbol
    if err := m.QueryRowsNoCacheCtx(ctx, &list, listSql, listArgs...); err != nil {
        return nil, 0, err
    }

    return list, total, nil
}
