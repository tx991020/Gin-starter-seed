-- +migrate Up
-- ----------------------------
-- Records of course
-- ----------------------------
ALTER TABLE `order` ADD COLUMN order_no varchar(128) NOT NULL;
ALTER TABLE `order` ADD  KEY order_no_idx(`order_no`);
ALTER TABLE `order` ADD  KEY  user_id_idx(`user_id`);
ALTER TABLE `order` ADD  KEY  ping_id_idx(`ping_id`);
-- +migrate Down