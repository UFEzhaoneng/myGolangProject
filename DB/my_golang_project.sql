CREATE database my_golang_project;
SET NAMES utf8;
USE my_golang_project;

CREATE TABLE IF NOT EXISTS `students` (
  `id` varchar(50) NOT NULL COMMENT '主键ID' primary key ,
  `name` varchar(30) NOT NULL COMMENT '姓名',
  `age` int NOT NULL COMMENT '年龄',
  `profession` varchar(30) NOT NULL COMMENT '专业',
  `created_at` datetime DEFAULT NULL COMMENT '创建时间',
  `updated_at` datetime DEFAULT NULL COMMENT '修改时间'
);