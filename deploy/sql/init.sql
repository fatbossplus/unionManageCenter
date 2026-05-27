-- ============================================================
-- 联盟管理中心 数据库初始化脚本
-- Database: union_manage
-- Charset: utf8mb4
-- ============================================================
-- 必须在连接建立后立即设置编码，防止中文注释乱码
SET NAMES utf8mb4;
SET character_set_client     = utf8mb4;
SET character_set_connection = utf8mb4;
SET character_set_results    = utf8mb4;

CREATE DATABASE IF NOT EXISTS `union_manage`
  DEFAULT CHARACTER SET utf8mb4
  DEFAULT COLLATE utf8mb4_unicode_ci;

USE `union_manage`;

-- ============================================================
-- 后台管理员表（登录本系统的运营/管理人员）
-- ============================================================
CREATE TABLE IF NOT EXISTS `admins` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `username`      VARCHAR(64)  NOT NULL COMMENT '登录用户名',
  `password`      VARCHAR(128) NOT NULL COMMENT '密码(bcrypt)',
  `email`         VARCHAR(128) NULL DEFAULT NULL COMMENT '邮箱',
  `phone`         VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '手机号',
  `real_name`     VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '真实姓名',
  `avatar`        VARCHAR(256) NOT NULL DEFAULT '' COMMENT '头像URL',
  `role_id`       BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '角色ID',
  `status`        TINYINT      NOT NULL DEFAULT 1 COMMENT '状态:1=正常,0=禁用',
  `last_login_at` DATETIME     NULL COMMENT '最后登录时间',
  `last_login_ip` VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '最后登录IP',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`    DATETIME     NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`),
  KEY `idx_role_id` (`role_id`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='后台管理员表';

-- ============================================================
-- 角色表
-- ============================================================
CREATE TABLE IF NOT EXISTS `roles` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name`        VARCHAR(64)  NOT NULL COMMENT '角色名称',
  `code`        VARCHAR(64)  NOT NULL COMMENT '角色编码(唯一)',
  `description` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '描述',
  `status`      TINYINT      NOT NULL DEFAULT 1 COMMENT '状态:1=启用,0=禁用',
  `sort`        INT          NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`  DATETIME     NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- ============================================================
-- 权限表（菜单/按钮/API 三合一）
-- ============================================================
CREATE TABLE IF NOT EXISTS `permissions` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `parent_id`   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '父ID,0=顶级',
  `name`        VARCHAR(64)  NOT NULL COMMENT '权限名称',
  `code`        VARCHAR(128) NOT NULL COMMENT '权限编码',
  `type`        TINYINT      NOT NULL DEFAULT 1 COMMENT '类型:1=菜单,2=按钮,3=API',
  `path`        VARCHAR(256) NOT NULL DEFAULT '' COMMENT '路由路径(前端)/API路径(后端)',
  `method`      VARCHAR(16)  NOT NULL DEFAULT '' COMMENT 'HTTP方法(GET/POST/PUT/DELETE)',
  `icon`        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '菜单图标',
  `component`   VARCHAR(128) NOT NULL DEFAULT '' COMMENT '前端组件路径',
  `sort`        INT          NOT NULL DEFAULT 0 COMMENT '排序',
  `visible`     TINYINT      NOT NULL DEFAULT 1 COMMENT '是否显示:1=显示,0=隐藏',
  `status`      TINYINT      NOT NULL DEFAULT 1 COMMENT '状态:1=启用,0=禁用',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`  DATETIME     NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_parent_id` (`parent_id`),
  KEY `idx_type` (`type`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';

-- ============================================================
-- 角色-权限关联
-- ============================================================
CREATE TABLE IF NOT EXISTS `role_permissions` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `role_id`       BIGINT UNSIGNED NOT NULL COMMENT '角色ID',
  `permission_id` BIGINT UNSIGNED NOT NULL COMMENT '权限ID',
  `created_at`    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_role_perm` (`role_id`, `permission_id`),
  KEY `idx_permission_id` (`permission_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联';

-- ============================================================
-- 平台用户表（联盟普通用户，非后台管理员）
-- ============================================================
CREATE TABLE IF NOT EXISTS `users` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `username`     VARCHAR(64)  NOT NULL COMMENT '用户名',
  `password`     VARCHAR(128) NOT NULL COMMENT '密码(bcrypt)',
  `email`        VARCHAR(128) NULL DEFAULT NULL COMMENT '邮箱',
  `phone`        VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '手机号',
  `real_name`    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '真实姓名',
  `avatar`       VARCHAR(256) NOT NULL DEFAULT '' COMMENT '头像URL',
  `status`       TINYINT      NOT NULL DEFAULT 1 COMMENT '状态:1=正常,2=待审核,3=禁用',
  `cert_status`  TINYINT      NOT NULL DEFAULT 0 COMMENT '实名认证:0=未认证,1=审核中,2=已认证',
  `source`       VARCHAR(32)  NOT NULL DEFAULT 'web' COMMENT '注册来源:web/mp_wx/app',
  `last_login_at` DATETIME    NULL COMMENT '最后登录时间',
  `last_login_ip` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '最后登录IP',
  `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`   DATETIME     NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`),
  KEY `idx_status` (`status`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='平台用户表';

-- ============================================================
-- 联盟表
-- ============================================================
CREATE TABLE IF NOT EXISTS `orgs` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name`          VARCHAR(128) NOT NULL COMMENT '联盟名称',
  `type`          VARCHAR(32)  NOT NULL DEFAULT 'ec' COMMENT '类型:ec=电商,service=服务,content=内容,other=其他',
  `description`   TEXT COMMENT '联盟简介',
  `logo`          VARCHAR(256) NOT NULL DEFAULT '' COMMENT 'Logo URL',
  `region`        VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '所在地区',
  `leader_id`     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '负责人用户ID',
  `contact_email` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '联系邮箱',
  `contact_phone` VARCHAR(20)  NOT NULL DEFAULT '' COMMENT '联系电话',
  `status`        TINYINT      NOT NULL DEFAULT 2 COMMENT '状态:1=正常,2=待审核,3=已冻结',
  `member_count`  INT          NOT NULL DEFAULT 0 COMMENT '成员数量(冗余)',
  `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`    DATETIME     NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_type` (`type`),
  KEY `idx_status` (`status`),
  KEY `idx_leader_id` (`leader_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='联盟表';

-- ============================================================
-- 联盟成员表
-- ============================================================
CREATE TABLE IF NOT EXISTS `org_members` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `org_id`     BIGINT UNSIGNED NOT NULL COMMENT '联盟ID',
  `user_id`    BIGINT UNSIGNED NOT NULL COMMENT '用户ID',
  `role`       VARCHAR(32) NOT NULL DEFAULT 'member' COMMENT '联盟内角色:owner/admin/member',
  `status`     TINYINT     NOT NULL DEFAULT 1 COMMENT '状态:1=正常,0=退出',
  `joined_at`  DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
  `created_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_org_user` (`org_id`, `user_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='联盟成员表';

-- ============================================================
-- 订单表
-- ============================================================
CREATE TABLE IF NOT EXISTS `orders` (
  `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `order_no`      VARCHAR(64)   NOT NULL COMMENT '订单号(唯一)',
  `type`          VARCHAR(32)   NOT NULL DEFAULT 'normal' COMMENT '类型:normal=普通,refund=退款',
  `status`        TINYINT       NOT NULL DEFAULT 1 COMMENT '状态:1=待支付,2=已支付,3=已退款,4=已取消',
  `pay_method`    VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '支付方式:wx/ali/bank',
  `amount`        DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '订单金额(元)',
  `user_id`       BIGINT UNSIGNED NOT NULL COMMENT '下单用户ID',
  `org_id`        BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '所属联盟ID',
  `remark`        VARCHAR(256)  NOT NULL DEFAULT '' COMMENT '备注',
  `paid_at`       DATETIME      NULL COMMENT '支付时间',
  `refunded_at`   DATETIME      NULL COMMENT '退款时间',
  `refund_reason` VARCHAR(256)  NOT NULL DEFAULT '' COMMENT '退款原因',
  `created_at`    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`    DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`    DATETIME      NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_org_id` (`org_id`),
  KEY `idx_status` (`status`),
  KEY `idx_created_at` (`created_at`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- ============================================================
-- 收款账户表
-- ============================================================
CREATE TABLE IF NOT EXISTS `finance_accounts` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `org_id`       BIGINT UNSIGNED NOT NULL COMMENT '联盟ID',
  `type`         VARCHAR(32)  NOT NULL DEFAULT 'bank' COMMENT '账户类型:bank=银行卡,ali=支付宝,wx=微信',
  `account_name` VARCHAR(128) NOT NULL COMMENT '账户名',
  `account_no`   VARCHAR(128) NOT NULL COMMENT '账号',
  `bank_name`    VARCHAR(128) NOT NULL DEFAULT '' COMMENT '开户行',
  `is_default`   TINYINT      NOT NULL DEFAULT 0 COMMENT '是否默认:1=是',
  `status`       TINYINT      NOT NULL DEFAULT 1 COMMENT '状态:1=正常,0=禁用',
  `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`   DATETIME     NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_org_id` (`org_id`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='收款账户表';

-- ============================================================
-- 财务结算表
-- ============================================================
CREATE TABLE IF NOT EXISTS `finance_settlements` (
  `id`           BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `org_id`       BIGINT UNSIGNED NOT NULL COMMENT '联盟ID',
  `account_id`   BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '收款账户ID',
  `amount`       DECIMAL(12,2) NOT NULL DEFAULT 0.00 COMMENT '结算金额(元)',
  `status`       TINYINT       NOT NULL DEFAULT 1 COMMENT '状态:1=待结算,2=结算中,3=已结算,4=失败',
  `period`       VARCHAR(32)   NOT NULL DEFAULT 'monthly' COMMENT '结算周期:daily/weekly/monthly',
  `period_start` DATE          NOT NULL COMMENT '结算周期起',
  `period_end`   DATE          NOT NULL COMMENT '结算周期止',
  `remark`       VARCHAR(256)  NOT NULL DEFAULT '' COMMENT '备注',
  `settled_at`   DATETIME      NULL COMMENT '实际结算时间',
  `operator_id`  BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作人ID',
  `created_at`   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`   DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `deleted_at`   DATETIME      NULL COMMENT '软删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_org_id` (`org_id`),
  KEY `idx_status` (`status`),
  KEY `idx_period` (`period_start`, `period_end`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='财务结算表';

-- ============================================================
-- 消息通知表
-- ============================================================
CREATE TABLE IF NOT EXISTS `messages` (
  `id`         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`    BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '接收用户ID,0=全体',
  `title`      VARCHAR(128) NOT NULL COMMENT '消息标题',
  `content`    TEXT         NOT NULL COMMENT '消息内容',
  `type`       VARCHAR(32)  NOT NULL DEFAULT 'system' COMMENT '类型:system/order/finance/security',
  `is_read`    TINYINT      NOT NULL DEFAULT 0 COMMENT '是否已读:0=未读,1=已读',
  `ref_id`     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联业务ID',
  `ref_type`   VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '关联业务类型',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `read_at`    DATETIME     NULL COMMENT '阅读时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_is_read` (`is_read`),
  KEY `idx_type` (`type`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='消息通知表';

-- ============================================================
-- 数据字典类型表
-- ============================================================
CREATE TABLE IF NOT EXISTS `dict_types` (
  `id`          INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name`        VARCHAR(64)  NOT NULL COMMENT '类型名称',
  `code`        VARCHAR(64)  NOT NULL COMMENT '类型编码(唯一)',
  `description` VARCHAR(256) NOT NULL DEFAULT '' COMMENT '描述',
  `status`      TINYINT      NOT NULL DEFAULT 1 COMMENT '状态:1=启用,0=禁用',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_code` (`code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据字典类型';

-- ============================================================
-- 数据字典项表
-- ============================================================
CREATE TABLE IF NOT EXISTS `dict_items` (
  `id`         INT UNSIGNED NOT NULL AUTO_INCREMENT,
  `type_id`    INT UNSIGNED NOT NULL COMMENT '字典类型ID',
  `label`      VARCHAR(128) NOT NULL COMMENT '显示名称',
  `value`      VARCHAR(128) NOT NULL COMMENT '字典值',
  `sort`       INT          NOT NULL DEFAULT 0 COMMENT '排序',
  `status`     TINYINT      NOT NULL DEFAULT 1 COMMENT '状态:1=启用,0=禁用',
  `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_type_id` (`type_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='数据字典项';

-- ============================================================
-- 操作日志表
-- ============================================================
CREATE TABLE IF NOT EXISTS `operation_logs` (
  `id`          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id`     BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '操作用户ID',
  `username`    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '操作用户名(冗余)',
  `action`      VARCHAR(64)  NOT NULL COMMENT '操作动作(create/update/delete/login等)',
  `resource`    VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '资源类型',
  `resource_id` BIGINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '资源ID',
  `detail`      TEXT COMMENT '操作详情(JSON)',
  `ip`          VARCHAR(64)  NOT NULL DEFAULT '' COMMENT '操作IP',
  `user_agent`  VARCHAR(256) NOT NULL DEFAULT '' COMMENT 'UA',
  `status`      TINYINT      NOT NULL DEFAULT 1 COMMENT '1=成功,0=失败',
  `created_at`  DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_action` (`action`),
  KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='操作日志';

-- ============================================================
-- 初始化种子数据
-- ============================================================

-- 角色
INSERT INTO `roles` (`id`, `name`, `code`, `description`, `status`, `sort`) VALUES
(1, '超级管理员', 'superadmin', '拥有所有权限',         1, 0),
(2, '联盟管理员', 'org_admin',  '管理本联盟内所有资源', 1, 1),
(3, '财务人员',   'finance',    '查看和处理财务结算',   1, 2),
(4, '运营人员',   'operator',   '日常运营操作',         1, 3);

-- 后台管理员（密码: admin123）
INSERT INTO `admins` (`id`, `username`, `password`, `email`, `real_name`, `role_id`, `status`) VALUES
(1, 'admin', '$2a$10$ow4orOc2sKz2DhStnPcILuorhEcGB3CpEkkb9p0PW5jKhltrRTj0S', 'admin@union.com', '超级管理员', 1, 1);

-- 菜单权限树
INSERT INTO `permissions` (`id`, `parent_id`, `name`, `code`, `type`, `path`, `icon`, `sort`, `visible`) VALUES
-- 一级菜单
(1,  0, '首页大屏', 'dashboard',   1, '/pages/dashboard/index',  '🏠', 0, 1),
(2,  0, '用户管理', 'user',        1, '/pages/users/index',       '👥', 1, 1),
(3,  0, '联盟管理', 'org',         1, '/pages/orgs/index',        '🏢', 2, 1),
(4,  0, '权限配置', 'permission',  1, '/pages/permissions/index', '🔐', 3, 1),
(5,  0, '订单中心', 'order',       1, '/pages/orders/index',      '📦', 4, 1),
(6,  0, '财务结算', 'finance',     1, '/pages/finance/index',     '💰', 5, 1),
(7,  0, '数据报表', 'report',      1, '/pages/reports/index',     '📊', 6, 1),
(8,  0, '消息通知', 'message',     1, '/pages/messages/index',    '💬', 7, 1),
(9,  0, '系统设置', 'settings',    1, '/pages/settings/index',    '⚙️', 8, 1),
-- 按钮权限
(10, 2, '用户新增', 'user:create', 2, '', '', 0, 0),
(11, 2, '用户编辑', 'user:update', 2, '', '', 1, 0),
(12, 2, '用户删除', 'user:delete', 2, '', '', 2, 0),
(14, 3, '联盟新增', 'org:create',  2, '', '', 0, 0),
(15, 3, '联盟编辑', 'org:update',  2, '', '', 1, 0),
(16, 3, '联盟删除', 'org:delete',  2, '', '', 2, 0),
(17, 5, '订单退款', 'order:refund',2, '', '', 0, 0),
(18, 6, '发起结算', 'finance:settle',2,'','', 0, 0);

-- 超级管理员拥有所有权限
INSERT INTO `role_permissions` (`role_id`, `permission_id`)
SELECT 1, id FROM `permissions`;

-- 数据字典
INSERT INTO `dict_types` (`name`, `code`, `description`) VALUES
('用户状态',  'user_status',   '用户账号状态'),
('联盟类型',  'org_type',      '联盟分类'),
('订单状态',  'order_status',  '订单状态枚举'),
('支付方式',  'pay_method',    '支付渠道'),
('结算状态',  'settle_status', '财务结算状态'),
('结算周期',  'settle_period', '财务结算周期');

INSERT INTO `dict_items` (`type_id`, `label`, `value`, `sort`) VALUES
(1, '正常',   '1', 0), (1, '待审核', '2', 1), (1, '已禁用', '3', 2),
(2, '电商联盟', 'ec', 0), (2, '服务联盟', 'service', 1), (2, '内容联盟', 'content', 2), (2, '其他', 'other', 3),
(3, '待支付', '1', 0), (3, '已支付', '2', 1), (3, '已退款', '3', 2), (3, '已取消', '4', 3),
(4, '微信支付', 'wx', 0), (4, '支付宝', 'ali', 1), (4, '银行卡', 'bank', 2),
(5, '待结算', '1', 0), (5, '结算中', '2', 1), (5, '已结算', '3', 2), (5, '失败',   '4', 3),
(6, '日结', 'daily', 0), (6, '周结', 'weekly', 1), (6, '月结', 'monthly', 2);
