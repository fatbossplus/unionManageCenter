-- ============================================================
-- 内容管理中心 (CMS) 建表语句
-- 字符集 utf8mb4，全表软删除，凭证 AES-256-GCM 加密
-- ============================================================

SET NAMES utf8mb4;
USE union_manage;

-- ----------------------------------------------------------
-- 1. 平台账号表（凭证密文存储，绝不明文落库）
-- ----------------------------------------------------------
CREATE TABLE IF NOT EXISTS cms_platform_accounts (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id       BIGINT UNSIGNED NOT NULL DEFAULT 0       COMMENT '归属联盟ID（0=全局）',
    platform     VARCHAR(20)  NOT NULL                    COMMENT 'wechat|rednote|douyin|csdn',
    account_name VARCHAR(128) NOT NULL                    COMMENT '账号显示名称',
    account_uid  VARCHAR(128) DEFAULT ''                  COMMENT '平台方唯一ID',
    cred_cipher  TEXT         NOT NULL                    COMMENT 'AES-256-GCM 凭证密文（Base64）',
    cred_iv      VARCHAR(64)  NOT NULL                    COMMENT 'GCM 初始化向量（Base64）',
    cred_version TINYINT      NOT NULL DEFAULT 1          COMMENT '密钥版本，支持密钥轮换',
    status       TINYINT      NOT NULL DEFAULT 1          COMMENT '0=禁用 1=正常 2=凭证失效 3=封禁',
    last_used_at DATETIME     DEFAULT NULL                COMMENT '最后使用时间',
    expires_at   DATETIME     DEFAULT NULL                COMMENT '凭证过期时间',
    remark       VARCHAR(256) DEFAULT ''                  COMMENT '备注',
    created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at   DATETIME     DEFAULT NULL,
    PRIMARY KEY (id),
    INDEX idx_org_platform (org_id, platform),
    INDEX idx_status       (status),
    INDEX idx_deleted      (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CMS平台账号';

-- ----------------------------------------------------------
-- 2. 驱动配置表（免费/付费切换，API Key 加密存储）
-- ----------------------------------------------------------
CREATE TABLE IF NOT EXISTS cms_driver_configs (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id       BIGINT UNSIGNED NOT NULL DEFAULT 0       COMMENT '归属联盟ID（0=全局默认）',
    config_key   VARCHAR(64)  NOT NULL                    COMMENT '配置键，如 wechat.scraper / ai.rewriter',
    driver_name  VARCHAR(64)  NOT NULL                    COMMENT '驱动名，如 csdn_rss_free / tongyi_paid',
    driver_type  VARCHAR(20)  NOT NULL DEFAULT 'free'     COMMENT 'free|paid',
    config_json  TEXT         DEFAULT NULL                COMMENT '驱动参数JSON（API Key等加密存储）',
    enabled      TINYINT      NOT NULL DEFAULT 1          COMMENT '是否启用',
    created_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_org_key (org_id, config_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CMS驱动配置';

-- ----------------------------------------------------------
-- 3. 采集任务表
-- ----------------------------------------------------------
CREATE TABLE IF NOT EXISTS cms_scrape_tasks (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id          BIGINT UNSIGNED NOT NULL DEFAULT 0    COMMENT '归属联盟',
    task_name       VARCHAR(128) NOT NULL                 COMMENT '任务名称',
    platform        VARCHAR(20)  NOT NULL                 COMMENT '采集平台',
    task_type       VARCHAR(20)  NOT NULL                 COMMENT 'search_title|follow_author',
    target_param    VARCHAR(512) NOT NULL                 COMMENT '搜索关键词 或 作者ID/URL',
    target_platform VARCHAR(20)  NOT NULL                 COMMENT '目标发布平台',
    account_id      BIGINT UNSIGNED DEFAULT 0             COMMENT '绑定的发布账号ID',
    cron_expr       VARCHAR(64)  DEFAULT ''               COMMENT 'Cron 表达式（空=手动）',
    fetch_limit     INT          NOT NULL DEFAULT 5       COMMENT '每次最多采集条数',
    status          TINYINT      NOT NULL DEFAULT 1       COMMENT '0=停用 1=启用',
    last_run_at     DATETIME     DEFAULT NULL,
    next_run_at     DATETIME     DEFAULT NULL,
    last_error      VARCHAR(512) DEFAULT '',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at      DATETIME     DEFAULT NULL,
    PRIMARY KEY (id),
    INDEX idx_org      (org_id),
    INDEX idx_platform (platform),
    INDEX idx_next_run (next_run_at),
    INDEX idx_deleted  (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CMS采集任务';

-- ----------------------------------------------------------
-- 4. 原始内容表（采集结果，去重）
-- ----------------------------------------------------------
CREATE TABLE IF NOT EXISTS cms_raw_contents (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    task_id         BIGINT UNSIGNED NOT NULL              COMMENT '来源任务ID',
    org_id          BIGINT UNSIGNED NOT NULL DEFAULT 0,
    platform        VARCHAR(20)  NOT NULL                 COMMENT '来源平台',
    source_url      VARCHAR(1024) DEFAULT ''              COMMENT '原文 URL',
    source_hash     VARCHAR(64)  NOT NULL                 COMMENT 'SHA256(source_url+title)，去重键',
    title           VARCHAR(512) DEFAULT ''               COMMENT '原文标题',
    author          VARCHAR(128) DEFAULT ''               COMMENT '作者',
    body_text       LONGTEXT     DEFAULT NULL             COMMENT '正文',
    body_images     JSON         DEFAULT NULL             COMMENT '图片URL列表',
    fetched_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    proc_status     VARCHAR(20)  NOT NULL DEFAULT 'pending'
                                                          COMMENT 'pending|processing|done|failed|skipped',
    proc_error      VARCHAR(512) DEFAULT '',
    created_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_hash    (source_hash),
    INDEX idx_task        (task_id),
    INDEX idx_proc_status (proc_status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CMS原始内容';

-- ----------------------------------------------------------
-- 5. AI 流水线处理记录（每轮独立存储，可追溯）
-- ----------------------------------------------------------
CREATE TABLE IF NOT EXISTS cms_process_records (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    raw_id       BIGINT UNSIGNED NOT NULL               COMMENT '关联原始内容',
    round        TINYINT NOT NULL                       COMMENT '处理轮次 1-5',
    round_name   VARCHAR(32) NOT NULL                   COMMENT 'clean|scan|rewrite|self_review|format',
    input_hash   VARCHAR(64) DEFAULT ''                 COMMENT '输入文本 SHA256（不存全文节省空间）',
    input_text   LONGTEXT DEFAULT NULL                  COMMENT '本轮输入',
    output_text  LONGTEXT DEFAULT NULL                  COMMENT '本轮输出',
    driver_used  VARCHAR(64) DEFAULT ''                 COMMENT '使用的驱动名',
    model_used   VARCHAR(64) DEFAULT ''                 COMMENT '使用的AI模型名',
    result       VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pass|fail|retry|pending',
    score_json   JSON DEFAULT NULL                      COMMENT '第4轮：6维度评分JSON',
    issues_json  JSON DEFAULT NULL                      COMMENT '发现的问题列表',
    retry_count  TINYINT NOT NULL DEFAULT 0,
    duration_ms  INT DEFAULT 0,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_raw_round (raw_id, round)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CMS AI流水线处理记录';

-- ----------------------------------------------------------
-- 6. 发布任务表
-- ----------------------------------------------------------
CREATE TABLE IF NOT EXISTS cms_publish_tasks (
    id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    org_id           BIGINT UNSIGNED NOT NULL DEFAULT 0,
    raw_id           BIGINT UNSIGNED NOT NULL             COMMENT '来源原始内容',
    account_id       BIGINT UNSIGNED NOT NULL             COMMENT '目标发布账号',
    target_platform  VARCHAR(20) NOT NULL                 COMMENT '目标平台',
    final_title      VARCHAR(512) DEFAULT ''              COMMENT '最终标题',
    final_text       LONGTEXT DEFAULT NULL                COMMENT '最终正文',
    final_images     JSON DEFAULT NULL                    COMMENT '最终图片列表',
    final_tags       JSON DEFAULT NULL                    COMMENT '标签列表',
    status           VARCHAR(20) NOT NULL DEFAULT 'draft' COMMENT 'draft|reviewing|approved|scheduled|published|failed',
    reviewed_by      BIGINT UNSIGNED DEFAULT 0            COMMENT '审核人（管理员ID）',
    reviewed_at      DATETIME DEFAULT NULL,
    scheduled_at     DATETIME DEFAULT NULL                COMMENT '定时发布时间',
    published_at     DATETIME DEFAULT NULL,
    platform_post_id VARCHAR(256) DEFAULT ''              COMMENT '发布后平台返回的文章ID',
    failure_reason   TEXT DEFAULT NULL,
    retry_count      TINYINT NOT NULL DEFAULT 0,
    created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at       DATETIME DEFAULT NULL,
    PRIMARY KEY (id),
    INDEX idx_org         (org_id),
    INDEX idx_account     (account_id),
    INDEX idx_status      (status),
    INDEX idx_scheduled   (scheduled_at),
    INDEX idx_deleted     (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CMS发布任务';

-- ----------------------------------------------------------
-- 7. 凭证访问审计日志（不可删除，只追加）
-- ----------------------------------------------------------
CREATE TABLE IF NOT EXISTS cms_credential_audit (
    id           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    account_id   BIGINT UNSIGNED NOT NULL               COMMENT '被访问的账号',
    operator_id  BIGINT UNSIGNED DEFAULT 0              COMMENT '操作的管理员ID（0=系统自动）',
    action       VARCHAR(32) NOT NULL                   COMMENT 'read|refresh|revoke|create|update',
    reason       VARCHAR(128) DEFAULT ''                COMMENT '操作原因',
    ip           VARCHAR(64) DEFAULT '',
    user_agent   VARCHAR(256) DEFAULT '',
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    INDEX idx_account (account_id),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='CMS凭证访问审计（只追加，不删除）';

-- ----------------------------------------------------------
-- 初始驱动配置（全局默认，org_id=0）
-- ----------------------------------------------------------
INSERT IGNORE INTO cms_driver_configs (org_id, config_key, driver_name, driver_type, enabled) VALUES
-- 采集驱动
(0, 'wechat.scraper',   'sogou_free',       'free', 1),
(0, 'rednote.scraper',  'playwright_free',  'free', 0),
(0, 'douyin.scraper',   'openapi_free',     'free', 0),
(0, 'csdn.scraper',     'rss_free',         'free', 1),
-- AI 改写驱动
(0, 'ai.rewriter',      'ollama_free',      'free', 1),
-- 合规检查驱动（可叠加，enabled 代表"启用该驱动加入检查链"）
(0, 'compliance.local', 'wordlist_free',    'free', 1),
(0, 'compliance.api',   'aliyun_paid',      'paid', 0),
-- 代理IP驱动
(0, 'proxy.pool',       'direct_free',      'free', 1),
-- 发布驱动
(0, 'wechat.publisher',   'official_api',    'free', 1),
(0, 'rednote.publisher',  'playwright_free', 'free', 0),
(0, 'douyin.publisher',   'openapi_free',    'free', 0),
(0, 'csdn.publisher',     'unofficial_api',  'free', 1);
