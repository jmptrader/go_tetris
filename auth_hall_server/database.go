package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gogames/go_tetris/types"
	"github.com/gogames/go_tetris/utils"
	"github.com/hprose/hprose-go/hprose"
)

const (
	sqlCreateAccounting = `CREATE TABLE accounting (
		txid VARCHAR(255),
		amount INT,
		address VARCHAR(255),
		account VARCHAR(255),
		isDeposit INT DEFAULT 0, -- 0 -> withdraw  1 -> deposit
		PRIMARY KEY (txid, isDeposit)
	) ENGINE=innoDB;`
	sqlCreateEnergy = `CREATE TABLE energy (
		uid INT,
		amount INT,
		created INT
	) ENGINE=innoDB;`
	sqlCreateSession = `CREATE TABLE sessions (
		sessionId VARCHAR(128),
		session BLOB,
		PRIMARY KEY (sessionId)
	) ENGINE=innoDB;`
)

var errDepositTxidExisted = fmt.Errorf("deposit txid exist")
var db *sql.DB

func initDatabase() {
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err.Error())
	}
	createTable()
	convFreezedToBalance()
	go keepDatabaseAlive()
	log.Info("initialize database...")
}

func createTable() {
	u := types.User{}
	for _, sql := range u.SqlGeneratorCreate() {
		if _, err := db.Exec(sql); err != nil {
			log.Debug("can not create user table: %v", err)
		}
	}
	if _, err := db.Exec(sqlCreateAccounting); err != nil {
		log.Debug("can not create accounting table: %v", err)
	}
	if _, err := db.Exec(sqlCreateEnergy); err != nil {
		log.Debug("can not create energy table: %v", err)
	}
	if _, err := db.Exec(sqlCreateSession); err != nil {
		log.Debug("can not create session table: %v", err)
	}
}

// init set bitcoin freezed to 0, add it to balance
func convFreezedToBalance() {
	_, err := db.Exec("UPDATE users SET Balance = Balance + Freezed, Freezed = 0;")
	if err != nil {
		panic("can not convert freezed to balance: " + err.Error())
	}
}

// ping database to keep connection alive
func keepDatabaseAlive() {
	defer utils.RecoverFromPanic("ping database panic: ", log.Critical, keepDatabaseAlive)
	for {
		time.Sleep(5 * time.Minute)
		db.Ping()
		db.Ping()
	}
}

// bitcoin withdraw
func insertWithdraw(txid, nickname, toAddr string, amount int) error {
	tx, err := db.Begin()
	defer func() {
		if err != nil {
			log.Error("can not insert into withdraw: %v", err)
			log.Error("txid -> %v\nnickname -> %v\ntoAddr -> %v\namount -> %vmBTC", txid, nickname, toAddr, amount)
		}
	}()
	if err != nil {
		return err
	}
	if err = func() error {
		if _, err := tx.Exec("INSERT INTO accounting(txid, amount, account, address, isDeposit) VALUES(?, ?, ?, ?, ?)",
			txid, amount, nickname, toAddr, 0); err != nil {
			return err
		}
		if _, err := tx.Exec("UPDATE users SET Balance = Balance - ? WHERE Nickname = ?", amount, nickname); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

// bitcoin deposit
func insertDeposit(txid, nickname, fromAddr string, amount int) error {
	// first check if it exists
	row := db.QueryRow("SELECT 1 FROM accounting WHERE isDeposit = 1 AND txid = ?", txid)
	if row.Scan(new(int)) != sql.ErrNoRows {
		return errDepositTxidExisted
	}
	// then insert
	tx, err := db.Begin()
	defer func() {
		if err != nil {
			log.Error("can not insert into deposit: %v", err)
			log.Error("txid -> %v\nnickname -> %v\nfromAddr -> %v\namount -> %vmBTC", txid, nickname, fromAddr, amount)
		}
	}()
	if err != nil {
		return err
	}
	if err = func() error {
		if _, err := tx.Exec("INSERT INTO accounting(txid, amount, account, address, isDeposit) VALUES(?, ?, ?, ?, ?)",
			txid, amount, nickname, fromAddr, 1); err != nil {
			return err
		}
		if _, err := tx.Exec("UPDATE users SET Balance = Balance + ? WHERE Nickname = ?", amount, nickname); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

// buy energy
func buyEnergy(uid, amount int) error {
	tx, err := db.Begin()
	defer func() {
		if err != nil {
			log.Error("buy energy get error: %v", err)
			log.Error("uid -> %v\namount of mBTC -> %v\n", uid, amount)
		}
	}()
	if err != nil {
		return err
	}
	if err = func() error {
		if _, err := tx.Exec("UPDATE users SET Balance = Balance - ?, Energy = Energy + ? WHERE Uid = ?", amount, amount*ratioEnergy2mBTC, uid); err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT INTO energy(uid, amount, created) VALUES(?, ?, ?)", uid, amount, time.Now().Unix()); err != nil {
			return err
		}
		return nil
	}(); err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	return err
}

// insert or update the users
func insertOrUpdateUser(us ...*types.User) {
	for _, u := range us {
		sql, args := u.SqlGeneratorUpdate()
		if _, err := db.Exec(sql, args...); err != nil {
			log.Error("can not update or insert -> error: %v\nsql: %v\nargs: %v\n", err, sql, args)
		}
	}
}

// store session into db before the program exit
func storeSession(sesses map[string]map[string]interface{}) {
	tx, err := db.Begin()
	if err != nil {
		log.Debug("can not start transaction for storing session: %v", err)
		return
	}
	defer func() {
		if err != nil {
			log.Debug("can not store session, get error: %v", err)
		} else {
			log.Info("successfully store all sessions in the database")
		}
	}()
	if err = func() error {
		for sessId, sess := range sesses {
			data, err := hprose.Marshal(sess)
			if err != nil {
				return err
			}
			if _, err := tx.Exec("INSERT INTO sessions VALUES(?, ?)", sessId, data); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		tx.Rollback()
		return
	}
	err = tx.Commit()
	return
}

// query all users
func queryUsers() []*types.User {
	sql := `SELECT * FROM users`
	rows, err := db.Query(sql)
	if err != nil {
		log.Error("can not query all users, error: %v", err)
		return nil
	}
	defer rows.Close()
	users := make([]*types.User, 0)
	for rows.Next() {
		u := &types.User{}
		if err := rows.Scan(&u.Uid, &u.Avatar, &u.Email, &u.Password, &u.Nickname,
			&u.Energy, &u.Level, &u.Win, &u.Lose, &u.Addr, &u.Balance, &u.Freezed,
			&u.Updated); err != nil {
			log.Error("can not scan user, error: %v", err)
			return nil
		}
		users = append(users, u)
	}
	return users
}

// query user by "field" & "val"
func queryUser(field string, val interface{}) *types.User {
	sql := fmt.Sprintf(`SELECT * FROM users WHERE %s = ?`, field)
	row := db.QueryRow(sql, val)
	u := &types.User{}
	if err := row.Scan(&u.Uid, &u.Avatar, &u.Email, &u.Password, &u.Nickname,
		&u.Energy, &u.Level, &u.Win, &u.Lose, &u.Addr, &u.Balance, &u.Freezed,
		&u.Updated); err != nil {
		log.Error("can not scan user, error: %v", err)
		return nil
	}
	return u
}

// query all sessions
func querySessions() map[string]map[string]interface{} {
	sql := `SELECT * FROM sessions`
	rows, err := db.Query(sql)
	if err != nil {
		log.Debug("can not query sessions: %v", err)
		return nil
	}
	defer rows.Close()
	res := make(map[string]map[string]interface{})
	for rows.Next() {
		var sessId string
		var sess []byte
		if err := rows.Scan(&sessId, &sess); err != nil {
			log.Debug("can not scan session: %v", err)
			return nil
		}
		var data = make(map[string]interface{})
		if err := hprose.Unmarshal(sess, &data); err != nil {
			log.Debug("can not unmarshal session: %v", err)
			return nil
		}
		res[sessId] = data
	}
	return res
}

// delete sessions
func deleteSessions() {
	if _, err := db.Exec("DELETE FROM sessions"); err != nil {
		log.Debug("can not delete all sessions in database: %v", err)
	}
}
