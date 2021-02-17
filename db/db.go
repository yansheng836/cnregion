// SPDX-License-Identifier: MIT

// Package db 提供区域数据库的相关操作
package db

import (
	"bytes"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/issue9/errwrap"
)

// DB 区域数据库信息
type DB struct {
	*Region
	versions []int // 支持的版本

	// 以下数据不会写入数据文件中

	fullNameSeparator string
}

// Load 从数据库文件加载数据
func Load(file, separator string) (*DB, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	return Unmarshal(data, separator)
}

// VersionIndex 指定年份在 Versions 中的下标
//
// 如果不存在，返回 -1
func (db *DB) VersionIndex(year int) int {
	for i, v := range db.versions {
		if v == year {
			return i
		}
	}
	return -1
}

// Dump 输出到文件
func (db *DB) Dump(file string) error {
	data, err := Marshal(db)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, data, os.ModePerm)
}

// Marshal 将 DB 转换成 []byte
func Marshal(db *DB) ([]byte, error) {
	return db.marshal()
}

// Unmarshal 解码 data 至 DB
func Unmarshal(data []byte, separator string) (*DB, error) {
	db := &DB{
		fullNameSeparator: separator,
	}
	if err := db.unmarshal(data); err != nil {
		return nil, err
	}
	return db, nil
}

// Find 查找指定 ID 对应的信息
func (db *DB) Find(id ...string) *Region {
	return db.findItem(id...)
}

func (db *DB) marshal() ([]byte, error) {
	vers := make([]string, 0, len(db.versions))
	for _, v := range db.versions {
		vers = append(vers, strconv.Itoa(v))
	}

	buf := errwrap.Buffer{Buffer: bytes.Buffer{}}
	buf.WByte('[')
	buf.WriteString(strings.Join(vers, ","))
	buf.WByte(']').WByte(':')

	err := db.Region.marshal(&buf)
	if err != nil {
		return nil, err
	}

	if buf.Err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (db *DB) unmarshal(data []byte) error {
	index := bytes.IndexByte(data, ':')
	arr := bytes.Trim(data[:index], "[]")
	arrs := strings.Split(string(arr), ",")
	db.versions = make([]int, 0, len(arrs))
	for _, item := range arrs {
		v, err := strconv.Atoi(item)
		if err != nil {
			return err
		}
		db.versions = append(db.versions, v)
	}

	data = data[index+1:]
	db.Region = &Region{}
	return db.Region.unmarshal(data, "", db.fullNameSeparator)
}
