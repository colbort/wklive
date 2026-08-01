# Scope 2 task integration fixture

`daily_reconciliation_scope2.sql` is intentionally destructive only to a newly created, isolated test database.
It seeds tenant 99 for `TestReconcileFullFundsMySQL`; the test then executes a healthy run, creates a wallet
mismatch, and repairs it.

From the repository root, create and load a fresh database in this order:

```sh
mysql -e 'CREATE DATABASE option_scope2_task_test'
mysql option_scope2_task_test < services/asset/asset.sql
mysql option_scope2_task_test < services/option/option.sql
mysql option_scope2_task_test < services/option/migrations/20260731_zt_option_daily_reconciliation_run.sql
mysql option_scope2_task_test < services/option/internal/logic/task/testdata/daily_reconciliation_scope2.sql
```

Then run from `services/option`:

```sh
OPTION_DAILY_RECONCILIATION_TASK_TEST_DSN='user:password@tcp(127.0.0.1:3306)/option_scope2_task_test?charset=utf8mb4&parseTime=true' \
  go test ./internal/logic/task -run TestReconcileFullFundsMySQL -count=1 -v
```

Drop the database after the test. Never point this fixture or test at a shared environment.
