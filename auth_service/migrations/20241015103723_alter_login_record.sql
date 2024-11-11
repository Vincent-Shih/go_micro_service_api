-- Modify "login_records" table
ALTER TABLE "login_records" ADD COLUMN "browser_ver" character varying NOT NULL, ADD COLUMN "platform" character varying NOT NULL, ADD COLUMN "country_code" character varying NOT NULL, ADD COLUMN "asp" character varying NOT NULL, ADD COLUMN "is_mobile" boolean NOT NULL, ADD COLUMN "err_message" character varying NOT NULL;
