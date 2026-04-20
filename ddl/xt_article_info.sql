-- phpMyAdmin SQL Dump
-- version 5.0.2
-- https://www.phpmyadmin.net/
--
-- 主机： 10.200.131.51
-- 生成日期： 2026-04-20 10:39:39
-- 服务器版本： 5.7.42-log
-- PHP 版本： 7.4.7

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- 数据库： `app.iqilu.com`
--

-- --------------------------------------------------------

--
-- 表的结构 `xt_article_info`
--

CREATE TABLE `xt_article_info` (
  `id` int(11) NOT NULL,
  `orgid` int(11) NOT NULL DEFAULT '0' COMMENT '机构id',
  `body` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `media` text COLLATE utf8_unicode_ci NOT NULL,
  `relate` text COLLATE utf8_unicode_ci NOT NULL COMMENT '相关推荐配置',
  `share_info` text COLLATE utf8_unicode_ci NOT NULL COMMENT '分享信息'
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

--
-- 转储表的索引
--

--
-- 表的索引 `xt_article_info`
--
ALTER TABLE `xt_article_info`
  ADD UNIQUE KEY `id_2` (`id`),
  ADD KEY `id` (`id`);
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
