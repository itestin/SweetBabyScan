package plugin_scan_weak

import (
	"database/sql"
	"fmt"
	_ "github.com/ClickHouse/clickhouse-go"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/sijms/go-ora/v2"
	"net"
	"strings"
	"time"
)

func CheckRDB(hostType, ip, user, pwd string, port uint) bool {
	flag := false

	switch hostType {
	case "mysql":
		flag = checkMySQL(ip, user, pwd, port)
	case "postgres":
		flag = checkPgSQL(ip, user, pwd, port)
	case "mssql":
		flag = checkMSSQL(ip, user, pwd, port)
	case "clickhouse":
		flag = checkClickHouse(ip, user, pwd, port)
	case "oracle":
		flag = checkOracle(ip, user, pwd, port)
	case "memcached":
		flag = checkMemcached(ip, user, pwd, port)
	}

	return flag
}

// 检测MySQL
func checkMySQL(ip, user, pwd string, port uint) bool {
	flag := false
	connStr := fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/?timeout=%v",
		user,
		pwd,
		ip,
		port,
		6*time.Second,
	)
	db, err := sql.Open("mysql", connStr)
	if err == nil {
		db.SetConnMaxLifetime(6 * time.Second)
		db.SetConnMaxIdleTime(6 * time.Second)
		db.SetMaxIdleConns(0)
		defer func() {
			db.Close()
		}()
		err = db.Ping()
		if err == nil {
			flag = true
		}
	}
	return flag
}

// 检测PgSQL
func checkPgSQL(ip, user, pwd string, port uint) bool {
	flag := false
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		user,
		pwd,
		ip,
		port,
		"postgres",
		"disable",
	)
	db, err := sql.Open("postgres", connStr)
	if err == nil {
		db.SetConnMaxLifetime(6 * time.Second)
		defer func() {
			db.Close()
		}()
		err = db.Ping()
		if err == nil {
			flag = true
		}
	}
	return flag
}

// 检测MSSQL
func checkMSSQL(ip, user, pwd string, port uint) bool {
	flag := false
	connStr := fmt.Sprintf(
		"server=%s;user id=%s;password=%s;port=%d;encrypt=disable;timeout=%v",
		ip,
		user,
		pwd,
		port,
		6*time.Second,
	)
	db, err := sql.Open("mssql", connStr)
	if err == nil {
		db.SetConnMaxLifetime(6 * time.Second)
		db.SetConnMaxIdleTime(6 * time.Second)
		db.SetMaxIdleConns(0)
		defer func() {
			db.Close()
		}()
		err = db.Ping()
		if err == nil {
			flag = true
		}
	}

	return flag
}

// 检测clickHouse
func checkClickHouse(ip, user, pwd string, port uint) bool {
	flag := false
	connStr := fmt.Sprintf(
		"tcp://%s:%d?username=%s&password=%s&read_timeout=6&write_timeout=6",
		ip,
		port,
		user,
		pwd,
	)
	db, err := sql.Open("clickhouse", connStr)
	if err == nil {
		defer func() {
			db.Close()
		}()
		err = db.Ping()
		if err == nil {
			flag = true
		}
	}

	return flag
}

// 检测oracle
func checkOracle(ip, user, pwd string, port uint) (status bool) {
	db, err := sql.Open(
		"oracle",
		fmt.Sprintf("oracle://%s:%s@%s:%d/orcl", user, pwd, ip, port),
	)
	if err == nil {
		db.SetConnMaxLifetime(6 * time.Second)
		db.SetConnMaxIdleTime(6 * time.Second)
		db.SetMaxIdleConns(0)
		defer db.Close()
		err = db.Ping()
		if err == nil {
			status = true
		}
	}
	return
}

// 检测memcached
func checkMemcached(ip, user, pwd string, port uint) (status bool) {
	client, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%v", ip, port), 6*time.Second)
	defer func() {
		if client != nil {
			client.Close()
		}
	}()
	if err == nil {
		err = client.SetDeadline(time.Now().Add(6 * time.Second))
		if err == nil {
			_, err = client.Write([]byte("stats\n"))
			if err == nil {
				rev := make([]byte, 1024)
				n, err := client.Read(rev)
				if err == nil {
					if strings.Contains(string(rev[:n]), "STAT") {
						status = true
					}
				}
			}
		}
	}
	return
}
