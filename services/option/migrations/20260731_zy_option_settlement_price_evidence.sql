-- Fail closed on malformed or unauditable settlement-price evidence.
-- Historical confirmed rows are not rewritten; operations metrics expose them.

SET @schema_name = DATABASE();
SET @ddl = IF(
  EXISTS(
    SELECT 1 FROM information_schema.statistics
    WHERE table_schema=@schema_name AND table_name='t_option_market_snapshot'
      AND index_name='idx_option_settlement_snapshot_evidence'
  ),
  'SELECT 1',
  'ALTER TABLE `t_option_market_snapshot` ADD KEY `idx_option_settlement_snapshot_evidence` (`tenant_id`,`contract_id`,`source_type`,`source_snapshot_id`,`snapshot_time`,`id`)'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

DROP TRIGGER IF EXISTS `trg_option_settlement_price_evidence_insert`;
DROP TRIGGER IF EXISTS `trg_option_settlement_price_evidence_update`;
DROP TRIGGER IF EXISTS `trg_option_settlement_price_evidence_delete`;

DELIMITER $$

CREATE TRIGGER `trg_option_settlement_price_evidence_insert`
BEFORE INSERT ON `t_option_settlement_price`
FOR EACH ROW
BEGIN
  DECLARE contract_count BIGINT DEFAULT 0;
  DECLARE expected_expire BIGINT DEFAULT 0;
  DECLARE expected_window BIGINT DEFAULT 0;
  DECLARE expected_min_samples BIGINT DEFAULT 0;
  DECLARE expected_source VARCHAR(32) DEFAULT '';
  DECLARE expected_method VARCHAR(32) DEFAULT '';
  DECLARE evidence_count BIGINT DEFAULT 0;
  DECLARE unique_count BIGINT DEFAULT 0;
  DECLARE blank_count BIGINT DEFAULT 0;
  DECLARE snapshot_match_count BIGINT DEFAULT 0;
  DECLARE expected_delivery DECIMAL(32,16) DEFAULT 0;

  IF NEW.status IN (1,2) THEN
    SELECT COUNT(1),COALESCE(MAX(expire_time),0),COALESCE(MAX(settlement_window_seconds),0),
           COALESCE(MAX(settlement_min_samples),0),COALESCE(MAX(settlement_price_source),''),
           COALESCE(MAX(settlement_price_method),'')
      INTO contract_count,expected_expire,expected_window,expected_min_samples,expected_source,expected_method
    FROM `t_option_contract`
    WHERE tenant_id=NEW.tenant_id AND id=NEW.contract_id;

    IF contract_count<>1 OR expected_expire<=0 OR expected_window<=0 OR expected_min_samples<=0 THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='settlement price contract rule is missing';
    END IF;
    IF NEW.window_start<>expected_expire-expected_window OR NEW.window_end<>expected_expire THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='settlement price window does not match contract';
    END IF;
    IF NEW.delivery_price<=0 OR NEW.sample_count<=0 OR JSON_VALID(NEW.source_snapshot_ids)=0
       OR JSON_TYPE(NEW.source_snapshot_ids)<>'ARRAY' THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='settlement price evidence is malformed';
    END IF;

    SELECT COUNT(1),COUNT(DISTINCT CAST(TRIM(evidence.source_id) AS BINARY)),
           COALESCE(SUM(TRIM(evidence.source_id)=''),0)
      INTO evidence_count,unique_count,blank_count
    FROM JSON_TABLE(
      NEW.source_snapshot_ids,
      '$[*]' COLUMNS(source_id VARCHAR(128) PATH '$')
    ) evidence;
    IF evidence_count<>NEW.sample_count OR unique_count<>evidence_count OR blank_count<>0 THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='settlement price evidence count, blank, or duplicate mismatch';
    END IF;

    IF BINARY NEW.price_source=BINARY expected_source
       AND BINARY NEW.calculation_method=BINARY expected_method THEN
      IF BINARY NEW.price_source<>BINARY 'authoritative-market'
         OR BINARY NEW.calculation_method<>BINARY 'MEDIAN'
         OR NEW.created_by<>0 OR NEW.sample_count<expected_min_samples THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='automatic settlement price rule is invalid';
      END IF;

      SELECT COUNT(1)
        INTO snapshot_match_count
      FROM JSON_TABLE(
        NEW.source_snapshot_ids,
        '$[*]' COLUMNS(source_id VARCHAR(128) PATH '$')
      ) evidence
      JOIN `t_option_market_snapshot` snapshot
        ON snapshot.tenant_id=NEW.tenant_id
       AND snapshot.contract_id=NEW.contract_id
       AND snapshot.source_type=1
       AND snapshot.snapshot_time BETWEEN NEW.window_start AND NEW.window_end
       AND snapshot.underlying_price>0
       AND BINARY TRIM(snapshot.source_snapshot_id)=BINARY TRIM(evidence.source_id);
      IF snapshot_match_count<>evidence_count THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='automatic settlement snapshot evidence is missing or duplicated';
      END IF;

      SELECT ROUND(AVG(ranked.underlying_price),16)
        INTO expected_delivery
      FROM (
        SELECT snapshot.underlying_price,
               ROW_NUMBER() OVER (ORDER BY snapshot.underlying_price,snapshot.id) row_no,
               COUNT(1) OVER () row_count
        FROM JSON_TABLE(
          NEW.source_snapshot_ids,
          '$[*]' COLUMNS(source_id VARCHAR(128) PATH '$')
        ) evidence
        JOIN `t_option_market_snapshot` snapshot
          ON snapshot.tenant_id=NEW.tenant_id
         AND snapshot.contract_id=NEW.contract_id
         AND snapshot.source_type=1
         AND snapshot.snapshot_time BETWEEN NEW.window_start AND NEW.window_end
         AND snapshot.underlying_price>0
         AND BINARY TRIM(snapshot.source_snapshot_id)=BINARY TRIM(evidence.source_id)
      ) ranked
      WHERE ranked.row_no IN (FLOOR((ranked.row_count+1)/2),FLOOR((ranked.row_count+2)/2));
      IF expected_delivery IS NULL OR NEW.delivery_price<>expected_delivery THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='automatic settlement price does not match snapshot median';
      END IF;
    ELSEIF BINARY NEW.price_source=BINARY 'manual-correction'
       AND BINARY NEW.calculation_method=BINARY 'MANUAL' THEN
      IF NEW.created_by<=0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='manual settlement correction requires a creator';
      END IF;
    ELSE
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='unsupported settlement price evidence type';
    END IF;

    IF NEW.status=2 AND (
      NEW.confirmed_by<=0 OR NEW.confirmed_at<=0
      OR (NEW.created_by>0 AND NEW.created_by=NEW.confirmed_by)
    ) THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='settlement price requires independent confirmation';
    END IF;
  END IF;
END$$

CREATE TRIGGER `trg_option_settlement_price_evidence_update`
BEFORE UPDATE ON `t_option_settlement_price`
FOR EACH ROW
BEGIN
  DECLARE contract_count BIGINT DEFAULT 0;
  DECLARE expected_expire BIGINT DEFAULT 0;
  DECLARE expected_window BIGINT DEFAULT 0;
  DECLARE expected_min_samples BIGINT DEFAULT 0;
  DECLARE expected_source VARCHAR(32) DEFAULT '';
  DECLARE expected_method VARCHAR(32) DEFAULT '';
  DECLARE evidence_count BIGINT DEFAULT 0;
  DECLARE unique_count BIGINT DEFAULT 0;
  DECLARE blank_count BIGINT DEFAULT 0;
  DECLARE snapshot_match_count BIGINT DEFAULT 0;
  DECLARE expected_delivery DECIMAL(32,16) DEFAULT 0;

  IF NEW.status IN (1,2) THEN
    SELECT COUNT(1),COALESCE(MAX(expire_time),0),COALESCE(MAX(settlement_window_seconds),0),
           COALESCE(MAX(settlement_min_samples),0),COALESCE(MAX(settlement_price_source),''),
           COALESCE(MAX(settlement_price_method),'')
      INTO contract_count,expected_expire,expected_window,expected_min_samples,expected_source,expected_method
    FROM `t_option_contract`
    WHERE tenant_id=NEW.tenant_id AND id=NEW.contract_id;

    IF contract_count<>1 OR expected_expire<=0 OR expected_window<=0 OR expected_min_samples<=0 THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='settlement price contract rule is missing';
    END IF;
    IF NEW.window_start<>expected_expire-expected_window OR NEW.window_end<>expected_expire THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='settlement price window does not match contract';
    END IF;
    IF NEW.delivery_price<=0 OR NEW.sample_count<=0 OR JSON_VALID(NEW.source_snapshot_ids)=0
       OR JSON_TYPE(NEW.source_snapshot_ids)<>'ARRAY' THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='settlement price evidence is malformed';
    END IF;

    SELECT COUNT(1),COUNT(DISTINCT CAST(TRIM(evidence.source_id) AS BINARY)),
           COALESCE(SUM(TRIM(evidence.source_id)=''),0)
      INTO evidence_count,unique_count,blank_count
    FROM JSON_TABLE(
      NEW.source_snapshot_ids,
      '$[*]' COLUMNS(source_id VARCHAR(128) PATH '$')
    ) evidence;
    IF evidence_count<>NEW.sample_count OR unique_count<>evidence_count OR blank_count<>0 THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='settlement price evidence count, blank, or duplicate mismatch';
    END IF;

    IF BINARY NEW.price_source=BINARY expected_source
       AND BINARY NEW.calculation_method=BINARY expected_method THEN
      IF BINARY NEW.price_source<>BINARY 'authoritative-market'
         OR BINARY NEW.calculation_method<>BINARY 'MEDIAN'
         OR NEW.created_by<>0 OR NEW.sample_count<expected_min_samples THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='automatic settlement price rule is invalid';
      END IF;

      SELECT COUNT(1)
        INTO snapshot_match_count
      FROM JSON_TABLE(
        NEW.source_snapshot_ids,
        '$[*]' COLUMNS(source_id VARCHAR(128) PATH '$')
      ) evidence
      JOIN `t_option_market_snapshot` snapshot
        ON snapshot.tenant_id=NEW.tenant_id
       AND snapshot.contract_id=NEW.contract_id
       AND snapshot.source_type=1
       AND snapshot.snapshot_time BETWEEN NEW.window_start AND NEW.window_end
       AND snapshot.underlying_price>0
       AND BINARY TRIM(snapshot.source_snapshot_id)=BINARY TRIM(evidence.source_id);
      IF snapshot_match_count<>evidence_count THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='automatic settlement snapshot evidence is missing or duplicated';
      END IF;

      SELECT ROUND(AVG(ranked.underlying_price),16)
        INTO expected_delivery
      FROM (
        SELECT snapshot.underlying_price,
               ROW_NUMBER() OVER (ORDER BY snapshot.underlying_price,snapshot.id) row_no,
               COUNT(1) OVER () row_count
        FROM JSON_TABLE(
          NEW.source_snapshot_ids,
          '$[*]' COLUMNS(source_id VARCHAR(128) PATH '$')
        ) evidence
        JOIN `t_option_market_snapshot` snapshot
          ON snapshot.tenant_id=NEW.tenant_id
         AND snapshot.contract_id=NEW.contract_id
         AND snapshot.source_type=1
         AND snapshot.snapshot_time BETWEEN NEW.window_start AND NEW.window_end
         AND snapshot.underlying_price>0
         AND BINARY TRIM(snapshot.source_snapshot_id)=BINARY TRIM(evidence.source_id)
      ) ranked
      WHERE ranked.row_no IN (FLOOR((ranked.row_count+1)/2),FLOOR((ranked.row_count+2)/2));
      IF expected_delivery IS NULL OR NEW.delivery_price<>expected_delivery THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='automatic settlement price does not match snapshot median';
      END IF;
    ELSEIF BINARY NEW.price_source=BINARY 'manual-correction'
       AND BINARY NEW.calculation_method=BINARY 'MANUAL' THEN
      IF NEW.created_by<=0 THEN
        SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='manual settlement correction requires a creator';
      END IF;
    ELSE
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='unsupported settlement price evidence type';
    END IF;

    IF NEW.status=2 AND (
      NEW.confirmed_by<=0 OR NEW.confirmed_at<=0
      OR (NEW.created_by>0 AND NEW.created_by=NEW.confirmed_by)
    ) THEN
      SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='settlement price requires independent confirmation';
    END IF;
  END IF;
END$$

CREATE TRIGGER `trg_option_settlement_price_evidence_delete`
BEFORE DELETE ON `t_option_settlement_price`
FOR EACH ROW
BEGIN
  SIGNAL SQLSTATE '45000' SET MESSAGE_TEXT='settlement price evidence cannot be deleted';
END$$

DELIMITER ;
