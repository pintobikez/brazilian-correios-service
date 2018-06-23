<<<<<<< HEAD
CREATE DATABASE correios;

USE correios;
=======
CREATE DATABASE correios_reverse;

USE correios_reverse;
>>>>>>> 6fd4253fa35eda9bb14e9c1e548abba73ac7caea

CREATE TABLE `request` (
  `request_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `request_type` varchar(10) NOT NULL,
  `request_service` varchar(10) NOT NULL,
  `colect_date` varchar(10) NULL,
  `order_nr` int(11) unsigned NOT NULL,
  `slip_number` varchar(12) NOT NULL,
  `origin_nome` varchar(60) NOT NULL,
  `origin_logradouro` varchar(72) NOT NULL,
  `origin_numero` int(11) NOT NULL,
  `origin_complemento` varchar(30) DEFAULT NULL,
  `origin_cep` varchar(8) NOT NULL,
  `origin_bairro` varchar(80) NOT NULL,
  `origin_cidade` varchar(40) NOT NULL,
  `origin_uf` varchar(2) NOT NULL,
  `origin_referencia` varchar(60) DEFAULT NULL,
  `origin_email` varchar(72) NOT NULL,
  `origin_ddd` varchar(4) DEFAULT '',
  `origin_telefone` varchar(12) DEFAULT '',
  `destination_nome` varchar(60) NOT NULL,
  `destination_logradouro` varchar(72) NOT NULL,
  `destination_numero` int(11) NOT NULL,
  `destination_complemento` varchar(30) DEFAULT NULL,
  `destination_cep` varchar(8) NOT NULL,
  `destination_bairro` varchar(80) NOT NULL,
  `destination_cidade` varchar(40) NOT NULL,
  `destination_uf` varchar(2) NOT NULL,
  `destination_referencia` varchar(60) DEFAULT NULL,
  `destination_email` varchar(72) NOT NULL,
  `callback` varchar(255) NOT NULL,
<<<<<<< HEAD
  `status` varchar(20) NOT NULL,
=======
  `status` varchar(10) NOT NULL,
>>>>>>> 6fd4253fa35eda9bb14e9c1e548abba73ac7caea
  `error_message` varchar(255) DEFAULT NULL,
  `retries` int(2) DEFAULT 0,
  `postage_code` varchar(10) DEFAULT NULL,
  `tracking_code` varchar(16) DEFAULT NULL,
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`request_id`),
  KEY `order_nr` (`order_nr`) USING BTREE,
  KEY `idx_postage_code` (`postage_code`) USING BTREE,
  KEY `idx_created_at` (`created_at`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8;

CREATE TABLE `request_item` (
  `order_item_id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `fk_request_id` int(11) unsigned NOT NULL,
  `item` varchar(64) NOT NULL,
  `product_name` varchar(255) NOT NULL,
  PRIMARY KEY (`order_item_id`),
  KEY `idx_request_id` (`fk_request_id`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8;
