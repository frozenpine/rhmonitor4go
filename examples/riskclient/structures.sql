/*
 Navicat Premium Data Transfer

 Source Server         : lingma
 Source Server Type    : SQLite
 Source Server Version : 3030001
 Source Schema         : trade

 Target Server Type    : SQLite
 Target Server Version : 3030001
 File Encoding         : 65001

 Date: 23/03/2023 21:53:47
*/

PRAGMA foreign_keys = false;

-- ----------------------------
-- Table structure for operation_account_kbar
-- ----------------------------
CREATE TABLE IF NOT EXISTS "operation_account_kbar" (
  "trading_day" DATE NOT NULL,
  "account_id" VARCHAR NOT NULL,
  "timestamp" TIMESTAMP NOT NULL,
  "duration" VARCHAR(255) NOT NULL,
  "open" FLOAT NOT NULL,
  "high" FLOAT NOT NULL,
  "low" FLOAT NOT NULL,
  "close" FLOAT NOT NULL,
  PRIMARY KEY ("trading_day", "account_id", "timestamp", "duration")
);

-- ----------------------------
-- Table structure for operation_trading_account
-- ----------------------------
CREATE TABLE IF NOT EXISTS "operation_trading_account" (
  "trading_day" DATE NOT NULL,
  "account_id" VARCHAR NOT NULL,
  "timestamp" TIMESTAMP NOT NULL,
  "pre_balance" FLOAT NOT NULL,
  "balance" FLOAT NOT NULL,
  "deposit" FLOAT NOT NULL,
  "withdraw" FLOAT NOT NULL,
  "profit" FLOAT NOT NULL,
  "fee" FLOAT NOT NULL,
  "margin" FLOAT NOT NULL,
  "available" FLOAT NOT NULL,
  PRIMARY KEY ("trading_day", "account_id", "timestamp")
);

PRAGMA foreign_keys = true;
