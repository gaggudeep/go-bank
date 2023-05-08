ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "owner_name_currency_key";

ALTER TABLE IF EXISTS "accounts" DROP CONSTRAINT IF EXISTS "accounts_owner_name_fkey";

DROP TABLE IF EXISTS "users";
