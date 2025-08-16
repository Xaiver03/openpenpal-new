CREATE TABLE `users` (`id` varchar(36),`username` varchar(50) NOT NULL,`email` varchar(100),`password_hash` varchar(255) NOT NULL,`nickname` varchar(50),`avatar` varchar(500),`role` varchar(20) NOT NULL DEFAULT "user",`school_code` varchar(20),`is_active` numeric DEFAULT true,`last_login_at` datetime,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime, `op_code` varchar(6),PRIMARY KEY (`id`));
CREATE INDEX `idx_users_deleted_at` ON `users`(`deleted_at`);
CREATE INDEX `idx_users_school_code` ON `users`(`school_code`);
CREATE UNIQUE INDEX `idx_users_email` ON `users`(`email`);
CREATE UNIQUE INDEX `idx_users_username` ON `users`(`username`);
CREATE TABLE `user_profiles` (`user_id` varchar(36),`real_name` varchar(50),`phone` varchar(20),`address` text,`bio` text,`preferences` json,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`user_id`),CONSTRAINT `fk_user_profiles_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`));
CREATE TABLE `envelopes` (`id` varchar(36),`design_id` varchar(36) NOT NULL,`user_id` varchar(36),`used_by` varchar(36),`letter_id` varchar(36),`barcode_id` varchar(100),`status` varchar(20) DEFAULT "unsent",`used_at` datetime,`recipient_op_code` varchar(6),`sender_op_code` varchar(6),`delivered_at` datetime,`tracking_info` json,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_envelopes_design` FOREIGN KEY (`design_id`) REFERENCES `envelope_designs`(`id`) ON DELETE SET NULL ON UPDATE CASCADE,CONSTRAINT `uni_envelopes_barcode_id` UNIQUE (`barcode_id`));
CREATE INDEX `idx_envelopes_sender_op_code` ON `envelopes`(`sender_op_code`);
CREATE INDEX `idx_envelopes_recipient_op_code` ON `envelopes`(`recipient_op_code`);
CREATE INDEX `idx_envelopes_letter_id` ON `envelopes`(`letter_id`);
CREATE TABLE `letters` (`id` varchar(36),`user_id` varchar(36) NOT NULL,`author_id` varchar(36) DEFAULT "",`title` varchar(255),`content` text NOT NULL,`style` varchar(20) NOT NULL DEFAULT "classic",`status` varchar(20) NOT NULL DEFAULT "draft",`visibility` varchar(20) NOT NULL DEFAULT "private",`like_count` integer DEFAULT 0,`recipient_op_code` varchar(6),`sender_op_code` varchar(6),`share_count` integer DEFAULT 0,`view_count` integer DEFAULT 0,`reply_to` varchar(36),`envelope_id` varchar(36),`created_at` datetime,`updated_at` datetime,`deleted_at` datetime, `type` varchar(20) NOT NULL DEFAULT "original", `metadata` jsonb, `recipient_id` varchar(36), `author_name` varchar(100), `scheduled_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_users_sent_letters` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),CONSTRAINT `fk_letters_envelope` FOREIGN KEY (`envelope_id`) REFERENCES `envelopes`(`id`),CONSTRAINT `fk_users_authored_letters` FOREIGN KEY (`author_id`) REFERENCES `users`(`id`));
CREATE INDEX `idx_letters_deleted_at` ON `letters`(`deleted_at`);
CREATE INDEX `idx_letters_envelope_id` ON `letters`(`envelope_id`);
CREATE INDEX `idx_letters_reply_to` ON `letters`(`reply_to`);
CREATE INDEX `idx_letters_sender_op_code` ON `letters`(`sender_op_code`);
CREATE INDEX `idx_letters_recipient_op_code` ON `letters`(`recipient_op_code`);
CREATE INDEX `idx_letters_author_id` ON `letters`(`author_id`);
CREATE INDEX `idx_letters_user_id` ON `letters`(`user_id`);
CREATE TABLE `letter_codes` (`id` varchar(36),`letter_id` varchar(36) NOT NULL,`code` varchar(50) NOT NULL,`qr_code_url` varchar(500),`qr_code_path` varchar(500),`expires_at` datetime,`created_at` datetime,`updated_at` datetime,`status` varchar(20) DEFAULT "unactivated",`recipient_code` varchar(6),`envelope_id` varchar(36),`bound_at` datetime,`delivered_at` datetime,`last_scanned_by` varchar(36),`last_scanned_at` datetime,`scan_count` integer DEFAULT 0,PRIMARY KEY (`id`),CONSTRAINT `fk_letter_codes_envelope` FOREIGN KEY (`envelope_id`) REFERENCES `envelopes`(`id`) ON DELETE SET NULL,CONSTRAINT `fk_letters_code` FOREIGN KEY (`letter_id`) REFERENCES `letters`(`id`) ON DELETE CASCADE);
CREATE INDEX `idx_letter_codes_envelope_id` ON `letter_codes`(`envelope_id`);
CREATE INDEX `idx_letter_codes_recipient_code` ON `letter_codes`(`recipient_code`);
CREATE INDEX `idx_letter_codes_status` ON `letter_codes`(`status`);
CREATE UNIQUE INDEX `idx_letter_codes_code` ON `letter_codes`(`code`);
CREATE UNIQUE INDEX `idx_letter_codes_letter_id` ON `letter_codes`(`letter_id`);
CREATE TABLE `status_logs` (`id` varchar(36),`letter_id` varchar(36) NOT NULL,`status` varchar(20) NOT NULL,`updated_by` varchar(36),`location` varchar(255),`note` text,`created_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_letters_status_logs` FOREIGN KEY (`letter_id`) REFERENCES `letters`(`id`) ON DELETE CASCADE);
CREATE INDEX `idx_status_logs_letter_id` ON `status_logs`(`letter_id`);
CREATE TABLE `letter_photos` (`id` varchar(36),`letter_id` varchar(36) NOT NULL,`image_url` varchar(500) NOT NULL,`is_public` numeric DEFAULT false,`created_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_letters_photos` FOREIGN KEY (`letter_id`) REFERENCES `letters`(`id`) ON DELETE CASCADE);
CREATE INDEX `idx_letter_photos_letter_id` ON `letter_photos`(`letter_id`);
CREATE TABLE `envelope_votes` (`id` varchar(36),`design_id` varchar(36) NOT NULL,`user_id` varchar(36) NOT NULL,`created_at` datetime,PRIMARY KEY (`id`));
CREATE TABLE `envelope_orders` (`id` varchar(36),`user_id` varchar(36) NOT NULL,`design_id` varchar(36) NOT NULL,`quantity` integer NOT NULL,`total_price` real NOT NULL,`status` varchar(20) DEFAULT "pending",`payment_method` varchar(50),`payment_id` varchar(100),`delivery_method` varchar(50),`delivery_info` json,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_envelope_orders_design` FOREIGN KEY (`design_id`) REFERENCES `envelope_designs`(`id`));
CREATE TABLE `ai_matches` (`id` varchar(36),`letter_id` varchar(36) NOT NULL,`matched_user_id` varchar(36),`match_score` real DEFAULT 0,`match_reason` text,`status` varchar(20) DEFAULT "pending",`provider` varchar(20),`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_ai_matches_matched_user_id` ON `ai_matches`(`matched_user_id`);
CREATE INDEX `idx_ai_matches_letter_id` ON `ai_matches`(`letter_id`);
CREATE TABLE `ai_replies` (`id` varchar(36),`original_letter_id` varchar(36) NOT NULL,`reply_letter_id` varchar(36),`persona` varchar(20) NOT NULL,`provider` varchar(20),`delay_hours` integer DEFAULT 24,`scheduled_at` datetime,`sent_at` datetime,`created_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_ai_replies_reply_letter_id` ON `ai_replies`(`reply_letter_id`);
CREATE INDEX `idx_ai_replies_original_letter_id` ON `ai_replies`(`original_letter_id`);
CREATE TABLE `ai_inspirations` (`id` varchar(36),`user_id` varchar(36),`theme` varchar(100),`prompt` text NOT NULL,`style` varchar(50),`tags` text,`usage_count` integer DEFAULT 0,`provider` varchar(20),`created_at` datetime,`is_active` numeric DEFAULT true,PRIMARY KEY (`id`));
CREATE INDEX `idx_ai_inspirations_user_id` ON `ai_inspirations`(`user_id`);
CREATE TABLE `ai_curations` (`id` varchar(36),`letter_id` varchar(36) NOT NULL,`exhibition_id` varchar(36),`category` varchar(50),`tags` text,`summary` text,`highlights` text,`score` real DEFAULT 0,`provider` varchar(20),`created_at` datetime,`approved_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_ai_curations_exhibition_id` ON `ai_curations`(`exhibition_id`);
CREATE INDEX `idx_ai_curations_letter_id` ON `ai_curations`(`letter_id`);
CREATE TABLE `ai_configs` (`id` varchar(36),`provider` varchar(20),`api_key` varchar(255),`api_endpoint` varchar(500),`model` varchar(100),`temperature` real DEFAULT 0.7,`max_tokens` integer DEFAULT 1000,`is_active` numeric DEFAULT true,`priority` integer DEFAULT 0,`daily_quota` integer DEFAULT 10000,`used_quota` integer DEFAULT 0,`quota_reset_at` datetime,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE UNIQUE INDEX `idx_ai_configs_provider` ON `ai_configs`(`provider`);
CREATE TABLE `ai_usage_logs` (`id` varchar(36),`user_id` varchar(36),`task_type` varchar(20) NOT NULL,`task_id` varchar(36),`provider` varchar(20),`model` varchar(100),`input_tokens` integer DEFAULT 0,`output_tokens` integer DEFAULT 0,`total_tokens` integer DEFAULT 0,`response_time` integer DEFAULT 0,`status` varchar(20),`error_message` text,`created_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_ai_usage_logs_task_id` ON `ai_usage_logs`(`task_id`);
CREATE INDEX `idx_ai_usage_logs_user_id` ON `ai_usage_logs`(`user_id`);
CREATE TABLE `user_credits` (`id` varchar(36),`user_id` text NOT NULL,`total` integer DEFAULT 0,`available` integer DEFAULT 0,`used` integer DEFAULT 0,`earned` integer DEFAULT 0,`level` integer DEFAULT 1,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE UNIQUE INDEX `idx_user_credits_user_id` ON `user_credits`(`user_id`);
CREATE TABLE `credit_transactions` (`id` varchar(36),`user_id` text NOT NULL,`type` text NOT NULL,`amount` integer NOT NULL,`description` text NOT NULL,`reference` text,`created_at` datetime, `expires_at` datetime, `expired_at` datetime, `is_expired` numeric DEFAULT false,PRIMARY KEY (`id`));
CREATE INDEX `idx_credit_transactions_user_id` ON `credit_transactions`(`user_id`);
CREATE TABLE `credit_rules` (`id` varchar(36),`action` text NOT NULL,`points` integer NOT NULL,`description` text NOT NULL,`is_active` numeric DEFAULT true,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE UNIQUE INDEX `idx_credit_rules_action` ON `credit_rules`(`action`);
CREATE TABLE `user_levels` (`id` varchar(36),`level` integer NOT NULL,`name` text NOT NULL,`required_exp` integer NOT NULL,`description` text,`benefits` text,`icon_url` text,`is_active` numeric DEFAULT true,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE UNIQUE INDEX `idx_user_levels_level` ON `user_levels`(`level`);
CREATE TABLE `museum_items` (`id` varchar(36),`source_type` varchar(20) NOT NULL,`source_id` varchar(36) NOT NULL,`title` varchar(200),`description` text,`tags` text,`status` varchar(20) DEFAULT "pending",`submitted_by` varchar(36),`approved_by` varchar(36),`approved_at` datetime,`view_count` integer DEFAULT 0,`like_count` integer DEFAULT 0,`share_count` integer DEFAULT 0,`comment_count` integer DEFAULT 0,`featured_at` datetime,`origin_op_code` varchar(6),`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_museum_items_approved_by_user` FOREIGN KEY (`approved_by`) REFERENCES `users`(`id`),CONSTRAINT `fk_museum_items_letter` FOREIGN KEY (`source_id`) REFERENCES `letters`(`id`),CONSTRAINT `fk_museum_items_submitted_by_user` FOREIGN KEY (`submitted_by`) REFERENCES `users`(`id`));
CREATE INDEX `idx_museum_items_origin_op_code` ON `museum_items`(`origin_op_code`);
CREATE TABLE `museum_collections` (`id` varchar(36),`name` varchar(200) NOT NULL,`description` text,`created_by` varchar(36) NOT NULL,`is_public` numeric DEFAULT true,`item_count` integer DEFAULT 0,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE TABLE `museum_exhibition_entries` (`id` varchar(36),`collection_id` varchar(36) NOT NULL,`item_id` varchar(36) NOT NULL,`display_order` integer DEFAULT 0,`created_at` datetime,PRIMARY KEY (`id`));
CREATE TABLE `museum_entries` (`id` varchar(36),`letter_id` varchar(36),`submission_id` varchar(36),`display_title` varchar(200),`author_display_type` varchar(20),`author_display_name` varchar(100),`curator_type` varchar(20),`curator_id` varchar(36),`categories` text[],`tags` text[],`status` varchar(20),`moderation_status` varchar(20),`view_count` integer DEFAULT 0,`like_count` integer DEFAULT 0,`bookmark_count` integer DEFAULT 0,`share_count` integer DEFAULT 0,`ai_metadata` text,`submitted_at` datetime,`approved_at` datetime,`featured_at` datetime,`withdrawn_at` datetime,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,PRIMARY KEY (`id`));
CREATE TABLE `museum_exhibitions` (`id` varchar(36),`title` varchar(200) NOT NULL,`description` text,`theme_keywords` text,`status` varchar(20) DEFAULT "draft",`creator_id` varchar(36),`start_date` datetime,`end_date` datetime,`max_entries` integer DEFAULT 50,`current_entries` integer DEFAULT 0,`view_count` integer DEFAULT 0,`is_public` numeric DEFAULT true,`is_featured` numeric DEFAULT false,`display_order` integer DEFAULT 0,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,PRIMARY KEY (`id`));
CREATE TABLE `moderation_records` (`id` varchar(36),`content_type` varchar(20) NOT NULL,`content_id` varchar(36) NOT NULL,`user_id` varchar(36) NOT NULL,`content` text,`image_urls` text,`status` varchar(20) NOT NULL DEFAULT "pending",`level` varchar(20),`score` real DEFAULT 0,`reasons` text,`categories` text,`a_iprovider` varchar(20),`ai_response` text,`reviewer_id` varchar(36),`review_note` text,`reviewed_at` datetime,`auto_moderated` numeric DEFAULT true,`created_at` datetime,`updated_at` datetime, `action` varchar(20), `reason` text, `notes` text,PRIMARY KEY (`id`));
CREATE INDEX `idx_moderation_records_user_id` ON `moderation_records`(`user_id`);
CREATE INDEX `idx_moderation_records_content_id` ON `moderation_records`(`content_id`);
CREATE INDEX `idx_moderation_records_content_type` ON `moderation_records`(`content_type`);
CREATE TABLE `sensitive_words` (`id` varchar(36),`word` varchar(100) NOT NULL,`category` varchar(50),`level` varchar(20) NOT NULL,`is_active` numeric DEFAULT true,`created_by` varchar(36),`created_at` datetime, `reason` text, `updated_at` datetime,PRIMARY KEY (`id`));
CREATE UNIQUE INDEX `idx_sensitive_words_word` ON `sensitive_words`(`word`);
CREATE TABLE `moderation_rules` (`id` varchar(36),`name` varchar(100) NOT NULL,`description` text,`content_type` varchar(20) NOT NULL,`rule_type` varchar(50),`pattern` text,`action` varchar(20),`priority` integer DEFAULT 0,`is_active` numeric DEFAULT true,`created_by` varchar(36),`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE TABLE `moderation_queues` (`id` varchar(36),`record_id` varchar(36) NOT NULL,`priority` integer DEFAULT 0,`assigned_to` varchar(36),`assigned_at` datetime,`status` varchar(20) DEFAULT "pending",`created_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_moderation_queues_record` FOREIGN KEY (`record_id`) REFERENCES `moderation_records`(`id`));
CREATE INDEX `idx_moderation_queues_record_id` ON `moderation_queues`(`record_id`);
CREATE TABLE `moderation_stats` (`id` varchar(36),`date` date NOT NULL,`content_type` varchar(20) NOT NULL,`total_count` integer DEFAULT 0,`approved_count` integer DEFAULT 0,`rejected_count` integer DEFAULT 0,`review_count` integer DEFAULT 0,`auto_moderate_count` integer DEFAULT 0,`avg_process_time` real DEFAULT 0,`created_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_moderation_stats_date` ON `moderation_stats`(`date`);
CREATE TABLE `notifications` (`id` varchar(36),`user_id` varchar(36) NOT NULL,`type` varchar(20) NOT NULL,`channel` varchar(20) NOT NULL,`priority` varchar(20) DEFAULT "normal",`title` varchar(200) NOT NULL,`content` text NOT NULL,`data` text,`status` varchar(20) DEFAULT "pending",`scheduled_at` datetime,`sent_at` datetime,`read_at` datetime,`retry_count` integer DEFAULT 0,`max_retries` integer DEFAULT 3,`error_message` text,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_notifications_user_id` ON `notifications`(`user_id`);
CREATE TABLE `email_templates` (`id` varchar(36),`name` varchar(100) NOT NULL,`type` varchar(20) NOT NULL,`subject` varchar(200) NOT NULL,`html_content` text NOT NULL,`plain_content` text,`variables` text,`is_active` numeric DEFAULT true,`priority` varchar(20) DEFAULT "normal",`created_by` varchar(36),`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE UNIQUE INDEX `idx_email_templates_name` ON `email_templates`(`name`);
CREATE TABLE `email_logs` (`id` varchar(36),`notification_id` varchar(36),`user_id` varchar(36),`to_email` varchar(255) NOT NULL,`from_email` varchar(255),`subject` varchar(500) NOT NULL,`template_id` varchar(36),`provider` varchar(50),`status` varchar(20) DEFAULT "pending",`sent_at` datetime,`delivered_at` datetime,`opened_at` datetime,`clicked_at` datetime,`bounced_at` datetime,`error_message` text,`retry_count` integer DEFAULT 0,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_email_logs_user_id` ON `email_logs`(`user_id`);
CREATE INDEX `idx_email_logs_notification_id` ON `email_logs`(`notification_id`);
CREATE TABLE `notification_preferences` (`id` varchar(36),`user_id` varchar(36) NOT NULL,`email_enabled` numeric DEFAULT true,`sms_enabled` numeric DEFAULT false,`push_enabled` numeric DEFAULT true,`types` text,`quiet_hours` varchar(50),`frequency` varchar(20) DEFAULT "realtime",`language` varchar(10) DEFAULT "zh-CN",`timezone` varchar(50) DEFAULT "Asia/Shanghai",`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE UNIQUE INDEX `idx_notification_preferences_user_id` ON `notification_preferences`(`user_id`);
CREATE TABLE `notification_batches` (`id` varchar(36),`name` varchar(200) NOT NULL,`type` varchar(20) NOT NULL,`channel` varchar(20) NOT NULL,`template_id` varchar(36),`target_users` text,`filter_conditions` text,`total_count` integer DEFAULT 0,`sent_count` integer DEFAULT 0,`failed_count` integer DEFAULT 0,`status` varchar(20) DEFAULT "preparing",`scheduled_at` datetime,`started_at` datetime,`completed_at` datetime,`created_by` varchar(36) NOT NULL,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE TABLE `analytics_metrics` (`id` varchar(36),`metric_type` varchar(20) NOT NULL,`metric_name` varchar(100) NOT NULL,`value` real NOT NULL,`unit` varchar(20),`dimension` varchar(100),`granularity` varchar(20) NOT NULL,`timestamp` datetime NOT NULL,`metadata` text,`created_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_analytics_metrics_timestamp` ON `analytics_metrics`(`timestamp`);
CREATE INDEX `idx_analytics_metrics_metric_type` ON `analytics_metrics`(`metric_type`);
CREATE TABLE `user_analytics` (`id` varchar(36),`user_id` varchar(36) NOT NULL,`date` date NOT NULL,`letters_sent` integer DEFAULT 0,`letters_received` integer DEFAULT 0,`letters_read` integer DEFAULT 0,`login_count` integer DEFAULT 0,`session_duration` integer DEFAULT 0,`courier_tasks` integer DEFAULT 0,`museum_visits` integer DEFAULT 0,`engagement_score` real DEFAULT 0,`retention_days` integer DEFAULT 0,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_user_analytics_date` ON `user_analytics`(`date`);
CREATE INDEX `idx_user_analytics_user_id` ON `user_analytics`(`user_id`);
CREATE TABLE `system_analytics` (`id` varchar(36),`date` date NOT NULL,`active_users` integer DEFAULT 0,`new_users` integer DEFAULT 0,`total_users` integer DEFAULT 0,`letters_created` integer DEFAULT 0,`letters_delivered` integer DEFAULT 0,`courier_tasks_completed` integer DEFAULT 0,`museum_items_added` integer DEFAULT 0,`avg_response_time` real DEFAULT 0,`error_rate` real DEFAULT 0,`server_uptime` real DEFAULT 0,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE UNIQUE INDEX `idx_system_analytics_date` ON `system_analytics`(`date`);
CREATE TABLE `performance_metrics` (`id` varchar(36),`endpoint` varchar(200) NOT NULL,`method` varchar(10) NOT NULL,`response_time` real NOT NULL,`status_code` integer NOT NULL,`user_agent` varchar(500),`ip_address` varchar(45),`user_id` varchar(36),`timestamp` datetime NOT NULL,`created_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_performance_metrics_timestamp` ON `performance_metrics`(`timestamp`);
CREATE INDEX `idx_performance_metrics_user_id` ON `performance_metrics`(`user_id`);
CREATE INDEX `idx_performance_metrics_endpoint` ON `performance_metrics`(`endpoint`);
CREATE TABLE `analytics_reports` (`id` varchar(36),`report_type` varchar(20) NOT NULL,`title` varchar(200) NOT NULL,`description` text,`granularity` varchar(20) NOT NULL,`start_date` datetime NOT NULL,`end_date` datetime NOT NULL,`data` text NOT NULL,`status` varchar(20) DEFAULT "generating",`generated_by` varchar(36),`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE TABLE `scheduled_tasks` (`id` text,`name` text NOT NULL,`description` text,`task_type` text NOT NULL,`priority` text DEFAULT "normal",`status` text DEFAULT "pending",`cron_expression` text,`scheduled_at` datetime,`next_run_at` datetime,`last_run_at` datetime,`last_status` text,`run_count` integer,`failure_count` integer,`payload` json,`max_retries` integer DEFAULT 3,`timeout_secs` integer DEFAULT 300,`is_active` numeric DEFAULT true,`start_date` datetime,`end_date` datetime,`max_runs` integer,`last_result` text,`last_error` text,`created_by` text,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_scheduled_tasks_deleted_at` ON `scheduled_tasks`(`deleted_at`);
CREATE TABLE `task_executions` (`id` text,`task_id` text NOT NULL,`status` text DEFAULT "pending",`started_at` datetime,`ended_at` datetime,`duration` integer,`result` text,`error` text,`output` text,`retry_count` integer DEFAULT 0,`worker_id` text,`server_host` text,`process_p_id` integer,`memory_usage` integer,`cpu_usage` real,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_task_executions_task` FOREIGN KEY (`task_id`) REFERENCES `scheduled_tasks`(`id`));
CREATE INDEX `idx_task_executions_task_id` ON `task_executions`(`task_id`);
CREATE TABLE `task_templates` (`id` text,`name` text NOT NULL,`description` text,`task_type` text NOT NULL,`priority` text DEFAULT "normal",`default_cron` text,`default_payload` json,`default_timeout` integer DEFAULT 300,`default_retries` integer DEFAULT 3,`is_enabled` numeric DEFAULT true,`category` text,`tags` text,`created_by` text,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE TABLE `task_workers` (`id` text,`name` text NOT NULL,`host` text NOT NULL,`port` integer,`status` text DEFAULT "active",`max_concurrency` integer DEFAULT 5,`current_tasks` integer DEFAULT 0,`completed_tasks` integer DEFAULT 0,`failed_tasks` integer DEFAULT 0,`last_heartbeat` datetime,`last_error` text,`supported_types` text,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE TABLE `storage_files` (`id` text,`file_name` text NOT NULL,`original_name` text NOT NULL,`file_size` integer NOT NULL,`mime_type` text,`extension` text,`category` text NOT NULL,`provider` text NOT NULL,`bucket_name` text,`object_key` text NOT NULL,`local_path` text,`public_url` text,`private_url` text,`thumbnail_url` text,`metadata` json,`status` text DEFAULT "active",`uploaded_by` text,`related_type` text,`related_id` text,`hash_md5` text,`hash_sha256` text,`is_public` numeric DEFAULT false,`access_token` text,`expires_at` datetime,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_storage_files_deleted_at` ON `storage_files`(`deleted_at`);
CREATE INDEX `idx_storage_files_related_id` ON `storage_files`(`related_id`);
CREATE INDEX `idx_storage_files_uploaded_by` ON `storage_files`(`uploaded_by`);
CREATE TABLE `storage_configs` (`id` text,`provider` text NOT NULL,`display_name` text NOT NULL,`config` json NOT NULL,`is_enabled` numeric DEFAULT false,`is_default` numeric DEFAULT false,`priority` integer DEFAULT 1,`max_file_size` integer DEFAULT 104857600,`max_total_size` integer,`current_size` integer DEFAULT 0,`allowed_types` text,`created_by` text,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `uni_storage_configs_provider` UNIQUE (`provider`));
CREATE INDEX `idx_storage_configs_deleted_at` ON `storage_configs`(`deleted_at`);
CREATE TABLE `storage_operations` (`id` text,`file_id` text NOT NULL,`operation` text NOT NULL,`user_id` text,`ip_address` text,`user_agent` text,`status` text,`bytes_transferred` integer,`duration` integer,`error_msg` text,`created_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_storage_operations_file` FOREIGN KEY (`file_id`) REFERENCES `storage_files`(`id`));
CREATE INDEX `idx_storage_operations_user_id` ON `storage_operations`(`user_id`);
CREATE INDEX `idx_storage_operations_file_id` ON `storage_operations`(`file_id`);
CREATE TABLE `museum_tags` (`id` varchar(36),`name` varchar(50) NOT NULL,`category` varchar(50) DEFAULT "general",`usage_count` integer DEFAULT 0,`created_at` datetime,PRIMARY KEY (`id`));
CREATE UNIQUE INDEX `idx_museum_tags_name` ON `museum_tags`(`name`);
CREATE TABLE `museum_interactions` (`id` varchar(36),`entry_id` varchar(36) NOT NULL,`user_id` varchar(36) NOT NULL,`type` varchar(20) NOT NULL,`created_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_museum_interactions_user_id` ON `museum_interactions`(`user_id`);
CREATE INDEX `idx_museum_interactions_entry_id` ON `museum_interactions`(`entry_id`);
CREATE TABLE `museum_reactions` (`id` varchar(36),`entry_id` varchar(36) NOT NULL,`user_id` varchar(36) NOT NULL,`reaction_type` varchar(20) NOT NULL,`comment` text,`created_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_museum_reactions_user_id` ON `museum_reactions`(`user_id`);
CREATE INDEX `idx_museum_reactions_entry_id` ON `museum_reactions`(`entry_id`);
CREATE TABLE `museum_submissions` (`id` varchar(36),`letter_id` varchar(36) NOT NULL,`submitted_by` varchar(36) NOT NULL,`display_preference` varchar(20) DEFAULT "anonymous",`pen_name` varchar(100),`submission_reason` text,`curator_notes` text,`status` varchar(20) DEFAULT "pending",`submitted_at` datetime,`reviewed_at` datetime,`reviewed_by` varchar(36),PRIMARY KEY (`id`),CONSTRAINT `fk_museum_submissions_letter` FOREIGN KEY (`letter_id`) REFERENCES `letters`(`id`));
CREATE TABLE `letter_templates` (`id` varchar(36),`name` varchar(255) NOT NULL,`description` text,`content` text DEFAULT "",`content_template` text,`style` varchar(20) NOT NULL DEFAULT "classic",`style_config` text,`category` varchar(100),`tags` varchar(500),`preview_image` varchar(500),`is_public` numeric DEFAULT true,`is_premium` numeric DEFAULT false,`is_active` numeric DEFAULT true,`usage_count` integer DEFAULT 0,`rating` real DEFAULT 0,`created_by` varchar(36),`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_letter_templates_creator` FOREIGN KEY (`created_by`) REFERENCES `users`(`id`) ON DELETE SET NULL);
CREATE INDEX `idx_letter_templates_created_by` ON `letter_templates`(`created_by`);
CREATE TABLE `letter_likes` (`id` varchar(36),`letter_id` varchar(36) NOT NULL,`user_id` varchar(36) NOT NULL,`created_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_letter_likes_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,CONSTRAINT `fk_letters_likes` FOREIGN KEY (`letter_id`) REFERENCES `letters`(`id`) ON DELETE CASCADE);
CREATE INDEX `idx_letter_likes_user_id` ON `letter_likes`(`user_id`);
CREATE INDEX `idx_letter_likes_letter_id` ON `letter_likes`(`letter_id`);
CREATE TABLE `letter_shares` (`id` varchar(36),`letter_id` varchar(36) NOT NULL,`user_id` varchar(36) NOT NULL,`platform` varchar(50),`share_url` varchar(500),`created_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_letter_shares_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE,CONSTRAINT `fk_letters_shares` FOREIGN KEY (`letter_id`) REFERENCES `letters`(`id`) ON DELETE CASCADE);
CREATE INDEX `idx_letter_shares_user_id` ON `letter_shares`(`user_id`);
CREATE INDEX `idx_letter_shares_letter_id` ON `letter_shares`(`letter_id`);
CREATE TABLE `user_daily_usages` (`id` varchar(36),`user_id` varchar(36) NOT NULL,`date` date NOT NULL,`inspirations_used` integer DEFAULT 0,`ai_replies_generated` integer DEFAULT 0,`penpal_matches` integer DEFAULT 0,`letters_curated` integer DEFAULT 0,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_user_daily_usages_date` ON `user_daily_usages`(`date`);
CREATE INDEX `idx_user_daily_usages_user_id` ON `user_daily_usages`(`user_id`);
CREATE TABLE `content_violation_records` (`id` varchar(36),`user_id` varchar(36) NOT NULL,`content_type` varchar(50) NOT NULL,`content_id` varchar(36),`original_text` text,`violation_type` varchar(100),`risk_level` varchar(20),`action` varchar(50),`review_status` varchar(20) DEFAULT "pending",`created_at` datetime,`reviewed_at` datetime,`reviewed_by` varchar(36),PRIMARY KEY (`id`));
CREATE INDEX `idx_content_violation_records_content_id` ON `content_violation_records`(`content_id`);
CREATE INDEX `idx_content_violation_records_user_id` ON `content_violation_records`(`user_id`);
CREATE INDEX `idx_users_op_code` ON `users`(`op_code`);
CREATE TABLE IF NOT EXISTS "couriers"  (`id` varchar(36),`user_id` varchar(36) NOT NULL,`name` text NOT NULL,`contact` text NOT NULL,`school` text NOT NULL,`zone` text NOT NULL,`managed_op_code_prefix` text,`has_printer` numeric DEFAULT false,`self_intro` text,`can_mentor` text DEFAULT "no",`weekly_hours` integer DEFAULT 5,`max_daily_tasks` integer DEFAULT 10,`transport_method` text,`time_slots` json,`status` text DEFAULT "pending",`level` integer DEFAULT 1,`task_count` integer DEFAULT 0,`points` integer DEFAULT 0,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,`zone_code` text,`zone_type` text,`parent_id` varchar(36),`created_by_id` varchar(36),`phone` text,`id_card` text,PRIMARY KEY (`id`),CONSTRAINT `fk_couriers_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),CONSTRAINT `fk_couriers_children` FOREIGN KEY (`parent_id`) REFERENCES `couriers`(`id`),CONSTRAINT `fk_couriers_created_by` FOREIGN KEY (`created_by_id`) REFERENCES `users`(`id`));
CREATE INDEX `idx_couriers_user_id` ON `couriers`(`user_id`);
CREATE INDEX `idx_couriers_managed_op_code_prefix` ON `couriers`(`managed_op_code_prefix`);
CREATE INDEX `idx_couriers_parent_id` ON `couriers`(`parent_id`);
CREATE INDEX `idx_couriers_deleted_at` ON `couriers`(`deleted_at`);
CREATE TABLE `level_upgrade_requests` (`id` integer PRIMARY KEY AUTOINCREMENT,`courier_id` varchar(36) NOT NULL,`current_level` integer NOT NULL,`request_level` integer NOT NULL,`reason` text,`evidence` text,`status` text DEFAULT "pending",`reviewed_by` varchar(36),`reviewed_at` datetime,`review_comment` text,`created_at` datetime,`updated_at` datetime,CONSTRAINT `fk_level_upgrade_requests_courier` FOREIGN KEY (`courier_id`) REFERENCES `couriers`(`id`),CONSTRAINT `fk_level_upgrade_requests_reviewer` FOREIGN KEY (`reviewed_by`) REFERENCES `users`(`id`));
CREATE TABLE sqlite_sequence(name,seq);
CREATE INDEX `idx_level_upgrade_requests_courier_id` ON `level_upgrade_requests`(`courier_id`);
CREATE TABLE `user_profiles_extended` (`user_id` varchar(36),`bio` text,`school` varchar(100),`op_code` varchar(6),`writing_level` integer DEFAULT 1,`courier_level` integer DEFAULT 0,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`user_id`),CONSTRAINT `fk_user_profiles_extended_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),CONSTRAINT `chk_user_profiles_extended_writing_level` CHECK (writing_level >= 0 AND writing_level <= 5),CONSTRAINT `chk_user_profiles_extended_courier_level` CHECK (courier_level >= 0 AND courier_level <= 4));
CREATE INDEX `idx_user_profiles_extended_op_code` ON `user_profiles_extended`(`op_code`);
CREATE TABLE `user_stats` (`user_id` varchar(36),`letters_sent` integer DEFAULT 0,`letters_received` integer DEFAULT 0,`museum_contributions` integer DEFAULT 0,`total_points` integer DEFAULT 0,`writing_points` integer DEFAULT 0,`courier_points` integer DEFAULT 0,`current_streak` integer DEFAULT 0,`max_streak` integer DEFAULT 0,`last_active_date` datetime,`updated_at` datetime,PRIMARY KEY (`user_id`),CONSTRAINT `fk_user_profiles_extended_stats` FOREIGN KEY (`user_id`) REFERENCES `user_profiles_extended`(`user_id`));
CREATE TABLE `user_privacy_settings` (`user_id` varchar(36),`show_email` numeric DEFAULT false,`show_op_code` numeric DEFAULT true,`show_stats` numeric DEFAULT true,`op_code_privacy` varchar(20) DEFAULT "partial",`profile_visible` numeric DEFAULT true,`updated_at` datetime,PRIMARY KEY (`user_id`),CONSTRAINT `fk_user_profiles_extended_privacy` FOREIGN KEY (`user_id`) REFERENCES `user_profiles_extended`(`user_id`));
CREATE TABLE `user_achievements` (`id` integer PRIMARY KEY AUTOINCREMENT,`user_id` varchar(36),`code` varchar(50),`name` varchar(100),`description` text,`icon` varchar(50),`category` varchar(50),`unlocked_at` datetime,CONSTRAINT `fk_user_profiles_extended_achievements` FOREIGN KEY (`user_id`) REFERENCES `user_profiles_extended`(`user_id`));
CREATE UNIQUE INDEX `idx_user_achievement` ON `user_achievements`(`code`);
CREATE INDEX `idx_user_achievements_user_id` ON `user_achievements`(`user_id`);
CREATE TABLE `scan_events` (`id` varchar(36),`barcode_id` varchar(36) NOT NULL,`letter_code_id` varchar(36) NOT NULL,`scanned_by` varchar(36) NOT NULL,`scan_type` varchar(20) NOT NULL,`location` varchar(255),`op_code` varchar(6),`latitude` real,`longitude` real,`old_status` varchar(20),`new_status` varchar(20),`device_info` text,`user_agent` text,`ip_address` varchar(45),`note` text,`metadata` jsonb,`timestamp` datetime NOT NULL,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_scan_events_letter_code` FOREIGN KEY (`letter_code_id`) REFERENCES `letter_codes`(`id`) ON DELETE CASCADE,CONSTRAINT `fk_scan_events_scanner` FOREIGN KEY (`scanned_by`) REFERENCES `users`(`id`) ON DELETE SET NULL);
CREATE INDEX `idx_scan_events_deleted_at` ON `scan_events`(`deleted_at`);
CREATE INDEX `idx_scan_events_timestamp` ON `scan_events`(`timestamp`);
CREATE INDEX `idx_scan_events_op_code` ON `scan_events`(`op_code`);
CREATE INDEX `idx_scan_events_scanned_by` ON `scan_events`(`scanned_by`);
CREATE INDEX `idx_scan_events_letter_code_id` ON `scan_events`(`letter_code_id`);
CREATE INDEX `idx_scan_events_barcode_id` ON `scan_events`(`barcode_id`);
CREATE TABLE op_code_schools (
    id VARCHAR(36) PRIMARY KEY,
    school_code VARCHAR(2) UNIQUE NOT NULL,
    school_name VARCHAR(100) NOT NULL,
    full_name VARCHAR(200),
    city VARCHAR(50),
    province VARCHAR(50),
    is_active BOOLEAN DEFAULT TRUE,
    managed_by VARCHAR(36), -- 四级信使ID
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_opcode_schools_school_code ON op_code_schools(school_code);
CREATE INDEX idx_opcode_schools_city ON op_code_schools(city);
CREATE INDEX idx_opcode_schools_province ON op_code_schools(province);
CREATE INDEX idx_opcode_schools_managed_by ON op_code_schools(managed_by);
CREATE TABLE op_code_areas (
    id VARCHAR(36) PRIMARY KEY,
    school_code VARCHAR(2) NOT NULL,
    area_code VARCHAR(2) NOT NULL,
    area_name VARCHAR(100) NOT NULL,
    description VARCHAR(200),
    is_active BOOLEAN DEFAULT TRUE,
    managed_by VARCHAR(36), -- 三级信使ID
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(school_code, area_code)
);
CREATE INDEX idx_opcode_areas_school_code ON op_code_areas(school_code);
CREATE INDEX idx_opcode_areas_area_code ON op_code_areas(area_code);
CREATE INDEX idx_opcode_areas_managed_by ON op_code_areas(managed_by);
CREATE TABLE op_codes (
    id VARCHAR(36) PRIMARY KEY,
    code VARCHAR(6) UNIQUE NOT NULL, -- 完整6位编码，如: PK5F3D
    school_code VARCHAR(2) NOT NULL, -- 前2位: 学校代码
    area_code VARCHAR(2) NOT NULL,   -- 中2位: 片区/楼栋代码
    point_code VARCHAR(2) NOT NULL,  -- 后2位: 具体位置代码
    
    -- 类型和属性
    point_type VARCHAR(20) NOT NULL, -- 类型: dormitory/shop/box/club
    point_name VARCHAR(100),         -- 位置名称
    full_address VARCHAR(200),       -- 完整地址描述
    is_public BOOLEAN DEFAULT FALSE, -- 后两位是否公开
    is_active BOOLEAN DEFAULT TRUE,  -- 是否激活
    
    -- 绑定信息
    binding_type VARCHAR(20),                    -- 绑定类型: user/shop/public
    binding_id VARCHAR(36),                      -- 绑定对象ID
    binding_status VARCHAR(20) DEFAULT 'pending', -- 绑定状态: pending/approved/rejected
    
    -- 管理信息
    managed_by VARCHAR(36) NOT NULL,  -- 管理者ID (二级信使)
    approved_by VARCHAR(36),          -- 审核者ID
    approved_at TIMESTAMP,            -- 审核时间
    
    -- 使用统计
    usage_count INTEGER DEFAULT 0,        -- 使用次数
    last_used_at TIMESTAMP,           -- 最后使用时间
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_opcodes_code ON op_codes(code);
CREATE INDEX idx_opcodes_school_code ON op_codes(school_code);
CREATE INDEX idx_opcodes_area_code ON op_codes(area_code);
CREATE INDEX idx_opcodes_point_code ON op_codes(point_code);
CREATE INDEX idx_opcodes_point_type ON op_codes(point_type);
CREATE INDEX idx_opcodes_is_active ON op_codes(is_active);
CREATE INDEX idx_opcodes_is_public ON op_codes(is_public);
CREATE INDEX idx_opcodes_managed_by ON op_codes(managed_by);
CREATE TABLE op_code_applications (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    requested_code VARCHAR(6), -- 申请的完整编码
    school_code VARCHAR(2) NOT NULL,
    area_code VARCHAR(2) NOT NULL,
    point_type VARCHAR(20) NOT NULL,
    point_name VARCHAR(100),
    full_address VARCHAR(200),
    reason TEXT,
    evidence TEXT, -- 证明材料JSON (SQLite doesn't have native JSON)
    
    status VARCHAR(20) DEFAULT 'pending', -- pending/approved/rejected
    assigned_code VARCHAR(6), -- 最终分配的编码
    reviewer_id VARCHAR(36),
    review_comment TEXT,
    reviewed_at TIMESTAMP,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_opcode_applications_user_id ON op_code_applications(user_id);
CREATE INDEX idx_opcode_applications_status ON op_code_applications(status);
CREATE INDEX idx_opcode_applications_school_code ON op_code_applications(school_code);
CREATE INDEX idx_opcode_applications_reviewer_id ON op_code_applications(reviewer_id);
CREATE TABLE op_code_permissions (
    id VARCHAR(36) PRIMARY KEY,
    courier_id VARCHAR(36) NOT NULL,
    courier_level INTEGER NOT NULL,
    code_prefix VARCHAR(6) NOT NULL, -- 管理的编码前缀
    permission VARCHAR(20) NOT NULL, -- view/assign/approve
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_opcode_permissions_courier_id ON op_code_permissions(courier_id);
CREATE INDEX idx_opcode_permissions_code_prefix ON op_code_permissions(code_prefix);
CREATE INDEX idx_opcode_permissions_courier_level ON op_code_permissions(courier_level);
CREATE VIEW v_opcode_full AS
SELECT 
    oc.id,
    oc.code,
    oc.school_code,
    os.school_name,
    os.city,
    os.province,
    oc.area_code,
    oa.area_name,
    oc.point_code,
    oc.point_type,
    oc.point_name,
    oc.full_address,
    oc.is_public,
    oc.is_active,
    oc.usage_count,
    oc.created_at,
    oc.updated_at
FROM op_codes oc
LEFT JOIN op_code_schools os ON oc.school_code = os.school_code
LEFT JOIN op_code_areas oa ON oc.school_code = oa.school_code AND oc.area_code = oa.area_code
WHERE oc.is_active = 1
/* v_opcode_full(id,code,school_code,school_name,city,province,area_code,area_name,point_code,point_type,point_name,full_address,is_public,is_active,usage_count,created_at,updated_at) */;
CREATE VIEW v_school_opcode_stats AS
SELECT 
    os.school_code,
    os.school_name,
    os.city,
    os.province,
    COUNT(oc.id) as total_opcodes,
    COUNT(CASE WHEN oc.is_active = 1 THEN 1 END) as active_opcodes,
    COUNT(CASE WHEN oc.is_public = 1 THEN 1 END) as public_opcodes,
    COUNT(CASE WHEN oc.point_type = 'dormitory' THEN 1 END) as dormitory_count,
    COUNT(CASE WHEN oc.point_type = 'shop' THEN 1 END) as shop_count,
    COUNT(CASE WHEN oc.point_type = 'box' THEN 1 END) as box_count,
    COUNT(CASE WHEN oc.point_type = 'club' THEN 1 END) as club_count
FROM op_code_schools os
LEFT JOIN op_codes oc ON os.school_code = oc.school_code
GROUP BY os.school_code, os.school_name, os.city, os.province
/* v_school_opcode_stats(school_code,school_name,city,province,total_opcodes,active_opcodes,public_opcodes,dormitory_count,shop_count,box_count,club_count) */;
CREATE INDEX `idx_letters_recipient_id` ON `letters`(`recipient_id`);
CREATE INDEX `idx_letters_scheduled_at` ON `letters`(`scheduled_at`);
CREATE TABLE IF NOT EXISTS "courier_tasks"  (`id` varchar(36),`courier_id` varchar(36) NOT NULL,`letter_code` varchar(50) NOT NULL,`title` varchar(200) NOT NULL,`sender_name` varchar(100) NOT NULL,`sender_phone` varchar(20),`recipient_hint` varchar(200),`target_location` varchar(200) NOT NULL,`current_location` varchar(200),`pickup_op_code` varchar(6),`delivery_op_code` varchar(6),`current_op_code` varchar(6),`priority` varchar(20) DEFAULT "normal",`status` varchar(20) DEFAULT "pending",`estimated_time` integer DEFAULT 30,`distance` decimal(10,2),`created_at` datetime,`updated_at` datetime,`deadline` datetime,`completed_at` datetime,`instructions` text,`reward` integer DEFAULT 10,`failure_reason` text,PRIMARY KEY (`id`),CONSTRAINT `fk_courier_tasks_courier` FOREIGN KEY (`courier_id`) REFERENCES `users`(`id`),CONSTRAINT `fk_courier_tasks_letter` FOREIGN KEY (`letter_code`) REFERENCES `letter_codes`(`code`));
CREATE INDEX `idx_courier_tasks_courier_id` ON `courier_tasks`(`courier_id`);
CREATE INDEX `idx_courier_tasks_letter_code` ON `courier_tasks`(`letter_code`);
CREATE INDEX `idx_courier_tasks_pickup_op_code` ON `courier_tasks`(`pickup_op_code`);
CREATE INDEX `idx_courier_tasks_delivery_op_code` ON `courier_tasks`(`delivery_op_code`);
CREATE INDEX `idx_courier_tasks_current_op_code` ON `courier_tasks`(`current_op_code`);
CREATE INDEX `idx_credit_transactions_expires_at` ON `credit_transactions`(`expires_at`);
CREATE INDEX `idx_credit_transactions_is_expired` ON `credit_transactions`(`is_expired`);
CREATE TABLE IF NOT EXISTS "envelope_designs"  (`id` varchar(36),`school_code` varchar(20),`type` varchar(20) DEFAULT "school",`theme` varchar(100),`image_url` varchar(500),`thumbnail_url` varchar(500),`creator_id` varchar(36) NOT NULL,`creator_name` varchar(100),`description` text,`status` varchar(20) DEFAULT "pending",`vote_count` integer DEFAULT 0,`period` varchar(50),`is_active` numeric DEFAULT true,`supported_op_code_prefix` varchar(4),`price` decimal(10,2) DEFAULT 3,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_envelope_designs_creator_id` ON `envelope_designs`(`creator_id`);
CREATE INDEX `idx_envelope_designs_supported_op_code_prefix` ON `envelope_designs`(`supported_op_code_prefix`);
CREATE INDEX `idx_envelope_designs_deleted_at` ON `envelope_designs`(`deleted_at`);
CREATE TABLE `cloud_personas` (`id` varchar(36),`user_id` varchar(36) NOT NULL,`name` varchar(100) NOT NULL,`relationship` varchar(50) NOT NULL,`description` text,`background_story` text,`personality` text,`memories` text,`last_interaction` datetime,`is_active` numeric DEFAULT true,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_cloud_personas_user_id` ON `cloud_personas`(`user_id`);
CREATE TABLE `cloud_letters` (`id` varchar(36),`user_id` varchar(36) NOT NULL,`persona_id` varchar(36) NOT NULL,`original_content` text NOT NULL,`ai_enhanced_draft` text,`final_content` text,`ai_reply` text,`status` varchar(20) DEFAULT "draft",`reviewer_level` integer DEFAULT 0,`reviewer_id` varchar(36),`review_comments` text,`delivery_date` datetime,`actual_delivery_date` datetime,`emotional_tone` varchar(50),`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_cloud_letters_persona_id` ON `cloud_letters`(`persona_id`);
CREATE INDEX `idx_cloud_letters_user_id` ON `cloud_letters`(`user_id`);
CREATE TABLE `credit_limit_rules` (`id` varchar(36),`action_type` text NOT NULL,`limit_type` text NOT NULL,`limit_period` text NOT NULL,`max_count` integer NOT NULL,`max_points` integer,`enabled` numeric DEFAULT true,`priority` integer DEFAULT 100,`description` text,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_credit_limit_rules_action_type` ON `credit_limit_rules`(`action_type`);
CREATE TABLE `user_credit_actions` (`id` varchar(36),`user_id` text NOT NULL,`action_type` text NOT NULL,`points` integer NOT NULL,`ip_address` varchar(45),`device_id` varchar(100),`user_agent` text,`reference` text,`created_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_user_credit_actions_reference` ON `user_credit_actions`(`reference`);
CREATE INDEX `idx_user_credit_actions_action_type` ON `user_credit_actions`(`action_type`);
CREATE INDEX `idx_user_credit_actions_user_id` ON `user_credit_actions`(`user_id`);
CREATE TABLE `credit_risk_users` (`user_id` varchar(36),`risk_score` decimal(5,2) DEFAULT 0,`risk_level` text DEFAULT "low",`blocked_until` datetime,`reason` text,`notes` text,`last_alert_at` datetime,`updated_at` datetime,`created_at` datetime,PRIMARY KEY (`user_id`));
CREATE TABLE `fraud_detection_logs` (`id` varchar(36),`user_id` text NOT NULL,`action_type` text NOT NULL,`risk_score` real NOT NULL,`is_anomalous` numeric NOT NULL,`detected_patterns` text,`evidence` text,`recommendations` text,`alert_count` integer DEFAULT 0,`created_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_fraud_detection_logs_user_id` ON `fraud_detection_logs`(`user_id`);
CREATE TABLE `credit_shop_categories` (`id` uuid,`name` varchar(100) NOT NULL,`description` text,`icon_url` text,`parent_id` uuid,`sort_order` integer DEFAULT 0,`is_active` numeric DEFAULT true,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_credit_shop_categories_children` FOREIGN KEY (`parent_id`) REFERENCES `credit_shop_categories`(`id`));
CREATE INDEX `idx_credit_shop_categories_parent_id` ON `credit_shop_categories`(`parent_id`);
CREATE TABLE `credit_shop_products` (`id` uuid,`name` varchar(200) NOT NULL,`description` text,`short_desc` varchar(500),`category` varchar(100),`product_type` varchar(50) NOT NULL,`credit_price` integer NOT NULL,`original_price` decimal(10,2),`stock` integer DEFAULT 0,`total_stock` integer DEFAULT 0,`redeem_count` integer DEFAULT 0,`image_url` text,`images` JSON,`tags` JSON,`specifications` JSON,`status` varchar(20) DEFAULT "active",`is_featured` numeric DEFAULT false,`is_limited` numeric DEFAULT false,`limit_per_user` integer DEFAULT 0,`priority` integer DEFAULT 0,`valid_from` datetime,`valid_to` datetime,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,PRIMARY KEY (`id`));
CREATE INDEX `idx_credit_shop_products_deleted_at` ON `credit_shop_products`(`deleted_at`);
CREATE TABLE `credit_carts` (`id` uuid,`user_id` varchar(36) NOT NULL,`total_items` integer DEFAULT 0,`total_credits` integer DEFAULT 0,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_credit_carts_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`));
CREATE INDEX `idx_credit_carts_user_id` ON `credit_carts`(`user_id`);
CREATE TABLE `credit_cart_items` (`id` uuid,`cart_id` uuid NOT NULL,`product_id` uuid NOT NULL,`quantity` integer NOT NULL DEFAULT 1,`credit_price` integer NOT NULL,`subtotal` integer NOT NULL,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_credit_cart_items_product` FOREIGN KEY (`product_id`) REFERENCES `credit_shop_products`(`id`),CONSTRAINT `fk_credit_carts_items` FOREIGN KEY (`cart_id`) REFERENCES `credit_carts`(`id`));
CREATE INDEX `idx_credit_cart_items_product_id` ON `credit_cart_items`(`product_id`);
CREATE INDEX `idx_credit_cart_items_cart_id` ON `credit_cart_items`(`cart_id`);
CREATE TABLE `credit_redemptions` (`id` uuid,`redemption_no` varchar(50) NOT NULL,`user_id` varchar(36) NOT NULL,`product_id` uuid NOT NULL,`quantity` integer NOT NULL DEFAULT 1,`credit_price` integer NOT NULL,`total_credits` integer NOT NULL,`status` varchar(20) DEFAULT "pending",`delivery_info` JSON,`redemption_code` varchar(100),`tracking_number` varchar(100),`notes` text,`processed_at` datetime,`shipped_at` datetime,`delivered_at` datetime,`completed_at` datetime,`cancelled_at` datetime,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_credit_redemptions_product` FOREIGN KEY (`product_id`) REFERENCES `credit_shop_products`(`id`),CONSTRAINT `fk_credit_redemptions_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`),CONSTRAINT `uni_credit_redemptions_redemption_no` UNIQUE (`redemption_no`));
CREATE INDEX `idx_credit_redemptions_product_id` ON `credit_redemptions`(`product_id`);
CREATE INDEX `idx_credit_redemptions_user_id` ON `credit_redemptions`(`user_id`);
CREATE TABLE `user_redemption_histories` (`id` uuid,`user_id` varchar(36) NOT NULL,`total_redemptions` integer DEFAULT 0,`total_credits_used` integer DEFAULT 0,`last_redemption_at` datetime,`favorite_category` varchar(100),`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`),CONSTRAINT `fk_user_redemption_histories_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`));
CREATE UNIQUE INDEX `idx_user_redemption_histories_user_id` ON `user_redemption_histories`(`user_id`);
CREATE TABLE `credit_shop_configs` (`id` uuid,`key` varchar(100) NOT NULL,`value` text NOT NULL,`description` text,`category` varchar(50),`is_editable` numeric DEFAULT true,`created_at` datetime,`updated_at` datetime,PRIMARY KEY (`id`));
CREATE UNIQUE INDEX `idx_credit_shop_configs_key` ON `credit_shop_configs`(`key`);
