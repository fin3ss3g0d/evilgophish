
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied
-- Move the relationship between campaigns and smtp to campaigns
ALTER TABLE campaigns ADD COLUMN "sms_id" bigint;
-- Create a new table to store smtp records
DROP TABLE sms;
CREATE TABLE sms(
	id integer primary key autoincrement,
	user_id bigint,
	name varchar(255),
    twilio_account_sid varchar(255),
    twilio_auth_token varchar(255),
    delay varchar(255),
    sms_from varchar(255),
	modified_date datetime default CURRENT_TIMESTAMP
);
-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

