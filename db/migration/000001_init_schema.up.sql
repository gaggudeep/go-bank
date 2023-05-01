CREATE TABLE "accounts" (
    "id" bigserial PRIMARY KEY,
    "owner_name" varchar NOT NULL,
    "balance" decimal NOT NULL CHECK("balance" >= 0),
    "currency" varchar NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "transactions" (
    "id" bigserial PRIMARY KEY,
    "account_id" bigserial NOT NULL,
    "amount" decimal NOT NULL CHECK("amount" != 0),
    "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "transfers" (
    "id" bigserial PRIMARY KEY,
    "from_account_id" bigserial NOT NULL,
    "to_account_id" bigserial NOT NULL,
    "amount" decimal NOT NULL CHECK("amount" > 0),
    "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE INDEX ON "accounts" ("owner_name");

CREATE INDEX ON "transactions" ("account_id");

CREATE INDEX ON "transfers" ("from_account_id");

CREATE INDEX ON "transfers" ("to_account_id");

CREATE INDEX ON "transfers" ("from_account_id", "to_account_id");

COMMENT ON COLUMN "transactions"."amount" IS 'must not be 0';

COMMENT ON COLUMN "transfers"."amount" IS 'must be positive';

ALTER TABLE "transactions" ADD FOREIGN KEY ("account_id") REFERENCES "accounts" ("id");

ALTER TABLE "transfers" ADD FOREIGN KEY ("from_account_id") REFERENCES "accounts" ("id");

ALTER TABLE "transfers" ADD FOREIGN KEY ("to_account_id") REFERENCES "accounts" ("id");
