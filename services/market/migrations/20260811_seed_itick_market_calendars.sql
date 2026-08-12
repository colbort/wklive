-- dbinit:baseline-safe
-- Generated from the official iTick product-list workbook downloaded 2026-08-11.
-- future-cn is intentionally omitted because the workbook does not publish its trading hours.
-- Index products are mapped only when the workbook also provides an explicit underlying-market schedule.
SET @now=CAST(UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3))*1000 AS UNSIGNED);
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('crypto','BA','','UTC',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='crypto' AND c.market='BA' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','00:00','24:00',0,127,1 FROM t_itick_market_calendar WHERE category_code='crypto' AND market='BA' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('forex','GB','','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='forex' AND c.market='GB' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','17:05','16:59',1,31,1 FROM t_itick_market_calendar WHERE category_code='forex' AND market='GB' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('forex','GB','ITICK_METAL','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='forex' AND c.market='GB' AND c.exchange='ITICK_METAL';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','18:05','16:59',1,31,1 FROM t_itick_market_calendar WHERE category_code='forex' AND market='GB' AND exchange='ITICK_METAL';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('forex','GB','ITICK_ENERGY','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='forex' AND c.market='GB' AND c.exchange='ITICK_ENERGY';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','20:00','17:00',1,31,1 FROM t_itick_market_calendar WHERE category_code='forex' AND market='GB' AND exchange='ITICK_ENERGY';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('forex','GB','ITICK_AGRI','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='forex' AND c.market='GB' AND c.exchange='ITICK_AGRI';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','20:00','14:19',1,31,1 FROM t_itick_market_calendar WHERE category_code='forex' AND market='GB' AND exchange='ITICK_AGRI';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','HK','','Asia/Hong_Kong',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='HK' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:30','12:00',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='HK' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','13:00','16:00',0,62,2 FROM t_itick_market_calendar WHERE category_code='stock' AND market='HK' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','US','','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='US' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'pre','04:00','09:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='US' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:30','16:00',0,62,2 FROM t_itick_market_calendar WHERE category_code='stock' AND market='US' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'post','16:00','20:00',0,62,3 FROM t_itick_market_calendar WHERE category_code='stock' AND market='US' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','SH','','Asia/Shanghai',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='SH' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:30','11:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='SH' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','13:00','15:00',0,62,2 FROM t_itick_market_calendar WHERE category_code='stock' AND market='SH' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','SZ','','Asia/Shanghai',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='SZ' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:30','11:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='SZ' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','13:00','15:00',0,62,2 FROM t_itick_market_calendar WHERE category_code='stock' AND market='SZ' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','TW','','Asia/Taipei',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='TW' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','13:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='TW' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','SG','','Asia/Singapore',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='SG' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','12:00',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='SG' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','13:00','17:00',0,62,2 FROM t_itick_market_calendar WHERE category_code='stock' AND market='SG' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','JP','','Asia/Tokyo',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='JP' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','11:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='JP' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','12:30','15:30',0,62,2 FROM t_itick_market_calendar WHERE category_code='stock' AND market='JP' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','TH','','Asia/Bangkok',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='TH' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','10:00','12:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='TH' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','14:00','16:30',0,62,2 FROM t_itick_market_calendar WHERE category_code='stock' AND market='TH' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','IN','','Asia/Kolkata',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='IN' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:15','15:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='IN' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','DE','','Europe/Berlin',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='DE' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','17:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='DE' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','MY','','Asia/Kuala_Lumpur',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='MY' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:30','12:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='MY' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','14:30','16:45',0,62,2 FROM t_itick_market_calendar WHERE category_code='stock' AND market='MY' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','TR','','Europe/Istanbul',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='TR' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','10:00','18:00',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='TR' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','ES','','Europe/Madrid',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='ES' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','08:30','17:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='ES' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','MX','','Etc/GMT+5',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='MX' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','08:30','15:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='MX' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','IT','','Europe/Rome',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='IT' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','17:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='IT' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','FR','','Europe/Paris',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='FR' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','17:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='FR' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','IL','','Asia/Jerusalem',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='IL' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:59','17:14',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='IL' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','NL','','Europe/Amsterdam',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='NL' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','17:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='NL' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','AR','','America/Argentina/Buenos_Aires',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='AR' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','11:00','17:00',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='AR' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','AU','','Australia/Sydney',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='AU' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:59','16:00',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='AU' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','VN','','Asia/Ho_Chi_Minh',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='VN' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','11:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='VN' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','13:00','15:00',0,62,2 FROM t_itick_market_calendar WHERE category_code='stock' AND market='VN' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','CA','','America/Toronto',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='CA' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:30','16:00',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='CA' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','PE','','America/Lima',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='PE' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','08:20','16:10',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='PE' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','GB','','Europe/London',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='GB' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','08:00','16:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='GB' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','ID','','Asia/Jakarta',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='ID' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','12:00',0,30,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='ID' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','13:30','15:00',0,30,2 FROM t_itick_market_calendar WHERE category_code='stock' AND market='ID' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','11:30',0,32,3 FROM t_itick_market_calendar WHERE category_code='stock' AND market='ID' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','14:00','15:00',0,32,4 FROM t_itick_market_calendar WHERE category_code='stock' AND market='ID' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','NG','','Africa/Lagos',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='NG' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','16:00',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='NG' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','PK','','Asia/Karachi',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='PK' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:30','15:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='PK' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','KE','','Africa/Nairobi',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='KE' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','15:00',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='KE' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','RO','','Europe/Bucharest',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='RO' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','10:00','18:00',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='RO' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','CH','','Europe/Zurich',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='CH' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','17:20',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='CH' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('stock','MA','','Africa/Casablanca',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='stock' AND c.market='MA' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:30','15:20',0,62,1 FROM t_itick_market_calendar WHERE category_code='stock' AND market='MA' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','HK','','Asia/Hong_Kong',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='HK' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:15','12:00',0,62,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='HK' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','13:00','16:30',0,62,2 FROM t_itick_market_calendar WHERE category_code='future' AND market='HK' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('fund','US','','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='fund' AND c.market='US' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'pre','04:00','09:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='fund' AND market='US' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:30','16:00',0,62,2 FROM t_itick_market_calendar WHERE category_code='fund' AND market='US' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'post','16:00','20:00',0,62,3 FROM t_itick_market_calendar WHERE category_code='fund' AND market='US' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('fund','SH','','Asia/Shanghai',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='fund' AND c.market='SH' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:30','11:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='fund' AND market='SH' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','13:00','15:00',0,62,2 FROM t_itick_market_calendar WHERE category_code='fund' AND market='SH' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('fund','SZ','','Asia/Shanghai',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='fund' AND c.market='SZ' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:30','11:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='fund' AND market='SZ' AND exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','13:00','15:00',0,62,2 FROM t_itick_market_calendar WHERE category_code='fund' AND market='SZ' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('indices','ES','','Europe/Madrid',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='indices' AND c.market='ES' AND c.exchange='';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','08:30','17:30',0,62,1 FROM t_itick_market_calendar WHERE category_code='indices' AND market='ES' AND exchange='';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_01','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_01';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','03:30','13:01',0,62,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_01';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_02','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_02';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','04:15','13:31',0,62,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_02';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_03','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_03';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','04:45','13:31',0,62,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_03';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_04','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_04';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','08:00','14:01',0,62,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_04';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_05','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_05';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','08:30','13:05',0,124,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_05';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_06','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_06';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','08:30','13:20',0,62,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_06';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_07','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_07';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','08:30','15:15',0,62,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_07';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_08','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_08';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','13:00',0,62,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_08';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_09','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_09';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','09:00','15:05',0,62,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_09';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_10','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_10';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','17:00','16:00',1,31,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_10';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_11','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_11';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','17:00','16:00',1,30,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_11';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','17:00','14:00',1,32,2 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_11';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_12','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_12';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','17:00','16:00',1,30,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_12';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','17:00','14:00',1,32,2 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_12';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_13','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_13';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','17:00','16:00',1,31,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_13';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_14','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_14';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','18:00','17:00',1,31,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_14';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_15','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_15';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','19:00','13:20',1,31,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_15';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_16','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_16';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','19:00','13:20',1,31,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_16';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_17','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_17';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','19:00','13:45',1,31,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_17';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_18','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_18';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','20:00','14:20',1,60,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_18';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','18:00','17:00',1,2,2 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_18';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_19','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_19';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','20:00','18:00',1,60,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_19';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','18:00','18:00',1,2,2 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_19';
INSERT INTO t_itick_market_calendar (category_code,market,exchange,timezone,trading_day_offset,week_start,enabled,remark,create_times,update_times)
VALUES ('future','US','ITICK_FUTURE_US_20','America/New_York',0,1,1,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE timezone=VALUES(timezone),enabled=1,remark=VALUES(remark),update_times=VALUES(update_times);
DELETE s FROM t_itick_market_session s JOIN t_itick_market_calendar c ON c.id=s.calendar_id WHERE c.category_code='future' AND c.market='US' AND c.exchange='ITICK_FUTURE_US_20';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','21:00','14:21',1,60,1 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_20';
INSERT INTO t_itick_market_session (calendar_id,session_type,start_time,end_time,cross_day,weekday_mask,sort)
SELECT id,'regular','18:00','14:21',1,2,2 FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_20';
DELETE FROM t_itick_product_calendar WHERE source='itick-product-list-2026-08-11';
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='forex' AND market='GB' AND exchange='ITICK_METAL' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('forex','GB','ALUMINUM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','LEAD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','NICKEL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAGCNY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAGEUR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAGHKD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAGSGD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAGUSD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAUAUD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAUCNH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAUCNY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAUEUR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAUGBP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAUHKD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAUSGD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAUTHB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAUUSD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XAUXAG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XCUUSD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XPDUSD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XPTUSD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','ZINC',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='forex' AND market='GB' AND exchange='ITICK_ENERGY' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('forex','GB','GASOLINE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','USOIL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XBRUSD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XNGUSD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','XTIUSD',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='forex' AND market='GB' AND exchange='ITICK_AGRI' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('forex','GB','COCOA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','COFFEE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','CORNF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','OATS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','SOYBEANS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','SOYMEAL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('forex','GB','SUGAR',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_01' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','SB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SB.Z',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_02' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','KC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KC.Z',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_03' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','CC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CC.Z',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_04' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','OJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','OJ.Z',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_05' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','BLK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PRK',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_06' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','CPO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CPV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','OPF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','POG',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_07' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','A2R',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AQR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ARR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ASR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BOS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CHI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CUS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DEN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EMV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ESR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EUS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EWE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EWM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EWW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FBT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FET',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FFV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IBHY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IBIG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IPC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IPCT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LAV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LAX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MAC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MIA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MIE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MIU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MMA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MMA.Y',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MXC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MXK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NYM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SDG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SFR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SON',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SR1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SR3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TBF3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TI3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TIE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VX44',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VX45',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VX47',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VX48',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VX49',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VXM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WDC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XBTF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XEU',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_08' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','SF',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_09' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','LBR',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_10' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','AUW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BAG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BAT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BEN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BET',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BGR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BGT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BLI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BLT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BME',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BMT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BPE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BPR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BPT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BST',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CCI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CCT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CVB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CWD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CWR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DFN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HRS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KWD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MFC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MFV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','OSF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QCW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RSO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SAS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SYP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UFB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UFE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UFV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UME',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UNO',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_11' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','DY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GDK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GNF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JQ',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_12' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','CB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CSC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DC',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_13' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','10Y',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','1E',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','1H',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','1NA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','1NM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','1OT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','1OZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','1S',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','1T',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','1ZA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','20U',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','22',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','23',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','25U',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','26',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','27',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','2C',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','2D',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','2FW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','2JW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','2YY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','2ZW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','30C',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','30J',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','30U',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','30Y',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','35U',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','3L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','3NA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','3NB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','3P',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','3V',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','3XW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','3ZW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','40U',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','45U',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','4C',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','4GC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','4V',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','4XW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','50U',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','51',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','55U',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','5L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','5Y',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','5YY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','60U',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','63',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','6A',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','6B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','6C',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','6E',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','6EP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','6J',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','6L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','6M',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','6N',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','6S',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','6Z',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','70U',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','7D',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','7F',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','7N',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','7V',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','7X',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','88',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','8D',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','8W',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','9Q',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A0D',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A0F',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A1D',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A1G',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A1L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A1M',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A1P',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A1Q',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A1R',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A1U',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A1V',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A1W',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A1X',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A32',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A33',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A38',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A3C',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A3G',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A3M',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A3N',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A3Q',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A3R',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A42',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A43',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A46',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A47',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A49',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A4L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A4M',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A4N',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A4P',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A4Q',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A4R',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A50',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A55',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A58',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A59',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A5C',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A6L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A6V',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A6W',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A6X',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A7E',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A7G',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A7I',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A7L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A7Q',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A7Y',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A81',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A8B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A8C',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A8G',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A8I',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A8J',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A8K',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A8L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A8M',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A8O',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A91',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','A9N',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AA3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AA4',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AA5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AA6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AA7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AA8',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AA9',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AB3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AB6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AB7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ABB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ABH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ABI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ABS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ABT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ABX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ABY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AC0',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ACB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ACD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ACS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ACU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AD0',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AD5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AD6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AD8',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AD9',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ADB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ADR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AE3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AE5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AE8',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AE9',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AEB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AEP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AET',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AEZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AF2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AF4',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AF5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AFE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AFF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AFH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AFI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AFK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AFR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AFY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AGA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AGE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AGT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AGX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AH3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AHB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AHJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AHL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AHM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AI1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AI2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AI3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AI4',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AI5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AI6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AI7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AI9',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AJ1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AJ2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AJ9',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AJB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AJJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AJL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AJR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AJS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AJY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AK1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AKL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AKR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AKS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AKX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AKZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AL1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AL5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AL6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AL8',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AL9',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ALA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ALB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ALI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ALM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ALX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ALY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AM1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AMB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AML',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AN1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ANE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ANL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ANT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AO1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AOB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AOH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AOJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AOL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AP1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AP2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AP3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AP4',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AP5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AP7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AP8',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AP9',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','APA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','APS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AQ5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AQ8',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AQA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AQK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AR0',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AR1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AR3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AR4',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AR6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AR8',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ARE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ARY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AS4',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AS7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ASD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ASI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ASN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ASP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ASPR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ASPT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AST',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AT0',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ATB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ATP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ATU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ATY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AU2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AU3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AU4',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AU5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AU6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AUB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AUF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AUH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AUI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AUJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AUP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AUS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AV0',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AVK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AVL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AVU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AVZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AW2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AW4',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AW6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AWJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AWN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AWQ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AX1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AXB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AY1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AYV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AYX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AZ0',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AZ1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AZ5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AZ7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','AZ9',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','B0',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','B1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','B1S',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','B2K',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','B6L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','B7H',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BAB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BCH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BEB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BEF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BFB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BFF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BFR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BG1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BG2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BG3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BHO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BIO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BKB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BKT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BLB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BOO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BPA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BPU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BR7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BTC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BTE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BUC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BUS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BWH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BZL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','BZS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','C2E',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','C3D',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','C4D',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','C4Z',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','C5D',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CAD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CBB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CC3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CCM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CFB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CFC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CGB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CHP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CJY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CLC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CLD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CLL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CLS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CMB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CMF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CMS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CNH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','COB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','COH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','COL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CPB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CPD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CPP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CRB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CRG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CS1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CS2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CS3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CS4',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CS5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CS6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CSX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CUP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CZK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','D1N',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','D2L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','D3L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','D4L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','D7L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DAB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DAX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DAZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DBB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DBL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DBT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DCB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DCL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DCW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DEB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DEP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DHA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DHB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DHY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DRS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DTF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DTH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DVE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','DX.Z',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','E1S',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','E3G',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','E4L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','E6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','E6M',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','E7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','E9X',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EAA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EAB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EAC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EAD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EAE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EAW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EBE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EBM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EBR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ECD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ECF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ECK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EDP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EEM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EFF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EFM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EGB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EGN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EHF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EHR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EHY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EIG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EJL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EL1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EMC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EMD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ENK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ENP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ENS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ENY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ENZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EO1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EOB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EP1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EPN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EPZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EQ1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ERL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ES',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ES1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ES2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ESB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ESG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ESK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ESQ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ESS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EST',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ESX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ETE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ETH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ETR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EU1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EU9',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EUB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EVC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EWB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EWF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EWG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EWN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EXR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','EZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','F1S',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','F3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FAL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FBD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FCB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FCN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FEF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FEW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FLB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FLJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FLP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FNG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FNG.Z',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FOA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FOM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FOR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FRC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FRS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FSF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FSS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FT1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FT5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FTL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FTU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','FVB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GBB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GBR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GCB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GCC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GCD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GCG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GCI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GCM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GCU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GDL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GEO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GES',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GFC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GIE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GKS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GMB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GMS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GNB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GNL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GNO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GNS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GOC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GSW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GUD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GWT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','GZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','H2L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','H2O',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','H5B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','H5F',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','H5G',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','H5L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HBX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HCS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HDG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HGB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HGS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HHT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HHW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HIA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HIL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HJC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HLT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HOA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HOB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HOL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HPD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HPE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HRC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HRP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HTA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HTB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HTT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HUF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HVG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HVO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HWA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HYB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','HYBT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IBS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IBV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IDL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IDR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ILS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IPO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IQB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IQBT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ITB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ITP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IUS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','IUT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','J4L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','J7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JBK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JBT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JCC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JCY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JDL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JET',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JFB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JFC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JKB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JKF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JKM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JKY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JNC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JNL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JPK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JPP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JPT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','JTB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','K2L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','K3L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','K4L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KAU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KEJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KEO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KEP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KGB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KKS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KMF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KMP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KOL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KRA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KRK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KRW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KRZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KSN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KSV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KZX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KZY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','L3L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LAF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LAP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LBU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LED',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LFA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LFG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LFM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LFU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LFW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LHV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LNG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LPE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LTC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','LTH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','M1B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','M2K',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','M35',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','M6A',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','M6B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','M6E',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MAA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MAB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MAE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MAF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MAS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MBA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MBB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MBC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MBE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MBL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MBM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MBO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MBR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MBS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MBT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MCB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MCD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MCE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MCF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MCL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MCN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MCS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MCT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MCU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MDB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MDD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ME',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MEB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MEE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MEF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MEO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MES',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MET',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MEU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MEW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MFB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MFD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MFP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MFR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MFS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MFU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MGB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MGC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MGE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MGF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MGH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MGN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MGS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MHE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MHG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MHO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MHT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MHY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MIH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MIN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MIP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MIR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MJB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MJC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MJN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MJP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MJY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MLE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MMC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MME',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MMF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MML',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MMM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MMN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MMO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MMP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MMR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MMW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MNB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MNC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MNG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MNH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MNI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MNK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MNQ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MNS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MNT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MOA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MOI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MOX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MPA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MPE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MPP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MPS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MPU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MPX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MQ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MQA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MRB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MRG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MRI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MRT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MSB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MSC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MSD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MSF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MSG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MSL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MST',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MT2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTQ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MTS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MUC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MUD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MUN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MUS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MUV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVA.Y',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVE.Y',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVH.Y',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MVV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MWA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MWA.Y',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MWL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MWN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MWS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MXB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MXP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MXR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MYM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MYY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','N1B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','N1S',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','N3L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','N3P',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','N9L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NA2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NA3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NAA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NBB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NBD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NBO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NBP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NCD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NCO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NCP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NDA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NEO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NEP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NFC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NFD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NFG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NFO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NGO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NHH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NHO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NHP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NIE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NIY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NJY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NKD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NLS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NMO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NMP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NNP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NOD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NOI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NOK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NOO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NOT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NPG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NQ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NQQ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NQT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NQX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NRO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NRP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NRR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NSK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NSO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NSP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NTP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NWM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NWO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NWP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NYF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NYP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','NZC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','OAD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','OFF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','OLB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','OMM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','OMN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ONB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','OOD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','OPO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PAC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PAD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PAL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PAM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PAU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PBO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PBP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PBT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PCD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PCO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PCP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PDJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PDL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PEL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PEX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PFO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PFP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PGG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PHF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PHO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PHP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PJY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PKP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PLE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PLM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PLN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PLO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PLP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PMF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PMO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PMP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PNF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PNK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PNL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','POB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','POC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','POL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PPL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PPP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PPW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PQO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PQP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PR4',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PR6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PRO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PRP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PSF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PSK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PSO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PSP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PTL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PTO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PUO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PUP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PVO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PVP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PWL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PXO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PXP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PYO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PYP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PZO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','PZP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QBTC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QCN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QCS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QDF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QDOW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QEF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QETH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QNDX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QNF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QOT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QRF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QRTY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QSF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QSPX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QTF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','QU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','R2G',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','R2V',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','R4B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','R53',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','R5B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','R5E',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','R5F',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','R5M',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','R5O',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','R6B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','R7L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RBB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RBF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RBL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RBM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RDA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RGF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RGI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RKA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RLB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RLX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RMB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RME',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RN3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RN4',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RN6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RNB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RS1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RSG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RSV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RTQ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RTX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RTY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RVR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','S1S',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','S53',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','S5B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','S5F',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','S5M',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','S5O',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SBM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SCB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SCT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SDA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SDD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SDI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SEK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SF1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SF3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SFB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SGB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SGC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SGD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SGF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SGO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SGU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SHR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SIL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SIR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SJY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SLB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SMC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SME',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SMET',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SMU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SNB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SOL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SOX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SPM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SR5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SRB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SRT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','STI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','STR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','STS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','STY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SXB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SXI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SXO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SXR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SXT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','SY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T1B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T1S',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T2B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T2D',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T2M',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T3B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T3L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T4B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T4D',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T5B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T5C',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T6B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T7C',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T7K',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T8B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','T8C',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TAS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TB2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TBK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TBM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TBT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TC1',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TC6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TC7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TCS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TD3',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TEF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TEM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TET',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TF2',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TFB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TFU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','THAI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','THB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','THD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','THG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TIL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TIO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TKB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TLB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TLD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TMB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TMD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TMS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TPD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TPY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TRB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TRI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TRL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TRM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TSL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TTB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TTD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TTE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TTF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TTG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TTH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TTI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TTP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TTS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TWE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','U7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','U9',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UCD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UCG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UCM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UCO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UCR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UCS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UDL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UHC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UHT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UKG',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ULB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UP5',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UPB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UPM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','USC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','USE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','USS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','UX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','V3L',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','V7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VDL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','VV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','W0',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WBR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WBX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WDB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WHB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WHD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WHT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WMB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WMD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WMR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WNB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WNT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WOL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WOW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WPL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WTB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WTD',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WTI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WTL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','WTT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','X0',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','X6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','X7',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','X9',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XAB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XAE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XAF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XAI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XAK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XAP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XAR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XAU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XAV',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XAY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XAZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XBT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XER',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XLB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XNB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XPP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XRP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XTT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XUB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XUK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XYT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YHE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YHF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YIA',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YIB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YIC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YID',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YIE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YII',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YIL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YIO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YIT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YIW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YIY',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YMX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YNO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YRP',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YRW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YUE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YVB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YWE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YWF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YWK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YZ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','Z1B',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','Z3N',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','Z4',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','Z6',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZAL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZB',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZF',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZGL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZJL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZKU',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZN',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZNC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZQ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZT',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_14' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','CJ',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','KT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','TT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','YO',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_15' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','KE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MZC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MZL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MZM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MZS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','MZW',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZL',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZO',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','ZW',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_16' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','ZS',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_17' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','MKC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XC',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XK',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','XW',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_18' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','RS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','RS.Z',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_19' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','AC',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='future' AND market='US' AND exchange='ITICK_FUTURE_US_20' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('future','US','CT',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('future','US','CT.Z',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='AU' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','AUS200',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','XJO',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='CA' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','CA60',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='CH' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','SWI20',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='DE' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','DAX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','GER30',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','GER40',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','MIDDE50',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','TECDE30',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='ES' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','ESP35',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','IBEX35',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','SPA35',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='FR' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','FR40EUR',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','FRA40',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='ID' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','COMPOSITE',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','IDX80',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','JII',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='IN' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','SENSEX',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='IT' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','IT40',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='JP' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','JP225',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','JPN225',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','JPXNIKKEI400',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','JPYBASKET',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','TOPIX',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='MX' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','ME',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='MY' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','FBM70',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','FBMKLCI',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','FBMT100',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='NL' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','NETH25',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='SG' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','FSTAS',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','STI',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
SET @calendar_id=(SELECT id FROM t_itick_market_calendar WHERE category_code='stock' AND market='SH' AND exchange='' LIMIT 1);
INSERT INTO t_itick_product_calendar (category_code,market,symbol,calendar_id,source,create_times,update_times) VALUES
('indices','GB','000001',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','000300',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','000680',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','000852',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','000905',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','399001',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','399006',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','ATMX',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','CHN50',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','CHNECOMM',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','CHNTECH',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','CSI300',@calendar_id,'itick-product-list-2026-08-11',@now,@now),
('indices','GB','XIN9',@calendar_id,'itick-product-list-2026-08-11',@now,@now)
ON DUPLICATE KEY UPDATE calendar_id=VALUES(calendar_id),source=VALUES(source),update_times=VALUES(update_times);
