-- noinspection SqlNoDataSourceInspectionForFile

CREATE SCHEMA IF NOT EXISTS pay;

CREATE TABLE pay.users (
   guid VARCHAR(255) PRIMARY KEY,
   gate_way VARCHAR(50) NOT NULL,
   account_id VARCHAR(50) NOT NULL,
   created_at TIMESTAMPTZ NOT NULL,
   updated_at TIMESTAMPTZ,
   CONSTRAINT unique_gateway_account UNIQUE (gate_way, account_id)
);


CREATE TABLE pay.transactions (
      transaction_id VARCHAR(255) PRIMARY KEY,
      type VARCHAR(50) NOT NULL,
      gate_way VARCHAR(50) NOT NULL,
      account_id VARCHAR(50) NOT NULL,
      user_id VARCHAR(255) NOT NULL,
      client_callback VARCHAR(255),
      amount DECIMAL(15, 2) NOT NULL,
      currency VARCHAR(10) NOT NULL,
      created_at TIMESTAMPTZ NOT NULL,
      updated_at TIMESTAMPTZ,
      status VARCHAR(50) NOT NULL,
      retry_count INT DEFAULT 0
);
