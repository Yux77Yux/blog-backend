package utils

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/yux77yux/blog-backend/config"
	"log"
)

func CreateTables() {
	db, err := sql.Open("mysql", "sa:x123@(192.168.101.4:3306)/")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS Blog")
	if err != nil {
		log.Fatal(err)
	}

	db, err = sql.Open("mysql", config.ConnectionStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	query := `
CREATE TABLE if not exists user(
	id INT AUTO_INCREMENT,
	username VARCHAR(50) NOT NULL,
	password VARCHAR(60) NOT NULL,

	PRIMARY KEY(id),
	UNIQUE INDEX index_user(username)
);
`
	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}

	query = `
CREATE TABLE if not exists user_incidental(
	uid char(9), -- 主键，主要用于被搜索，生成后不可更改
	id INT NOT NULL,
	name VARCHAR(100), -- 昵称，也可用于被搜索，允许可以更改
    bio TEXT,  -- 个性签名
    profile TEXT,  -- 头像图片地址
    status TINYINT(1) not null default 0,  -- 登录状态
    popularity DECIMAL(10,5), -- 主页受欢迎程度，用于优先搜索

	PRIMARY KEY(uid),
	FOREIGN KEY (id) REFERENCES user(id)
	 ON DELETE restrict 
	 ON UPDATE cascade
);
`
	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}

	query = `
CREATE TABLE if not exists article(
	uuid char(16), -- 主键，唯一搜索
    uid char(9) NOT NULL, -- 外键， user_incidental.Uid,可链接到作者页
    title TINYTEXT NOT NULL, -- 文章标题，可用于模糊搜索
    titleLight TINYTEXT NOT NULL, -- 标题是否高亮
    coverDimensions varchar(30) NOT NULL, -- 封面大小，用于类选择器
    coverImageUrl TEXT NOT NULL, -- 帖子封面
    summary TEXT NOT NULL, -- 文章概述
    content MEDIUMTEXT NOT NULL, -- 文章内容
    createdAt DATETIME NOT NULL, -- 创建时间
    updatedAt DATETIME NOT NULL, -- 更新时间
	timezone VARCHAR(50) NOT NULL, -- 发布时的时区
    views INT NOT NULL default 1, --  浏览次数
    likes INT NOT NULL default 0, --  点赞次数
    tags TEXT, -- 文章分类标签
    status TINYINT(1) NOT NULL DEFAULT 0,  -- 0草稿，1发布
    popularity DECIMAL(10,5) NOT NULL DEFAULT 1.0, -- 流行度

	PRIMARY KEY(uuid),
	FOREIGN KEY (uid) REFERENCES user_incidental(uid)
);
`
	if _, err := db.Exec(query); err != nil {
		log.Fatal(err)
	}
}
