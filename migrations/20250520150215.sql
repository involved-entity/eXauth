-- Create "users" table
CREATE TABLE "public"."users" ("id" bigserial NOT NULL, "joined_at" timestamptz NULL, "username" text NOT NULL, "password" text NOT NULL, "email" text NOT NULL, "is_verified" boolean NULL DEFAULT false, "is_admin" boolean NULL DEFAULT false, PRIMARY KEY ("id"));
-- Create index "idx_users_email" to table: "users"
CREATE UNIQUE INDEX "idx_users_email" ON "public"."users" ("email");
-- Create index "idx_users_username" to table: "users"
CREATE UNIQUE INDEX "idx_users_username" ON "public"."users" ("username");
