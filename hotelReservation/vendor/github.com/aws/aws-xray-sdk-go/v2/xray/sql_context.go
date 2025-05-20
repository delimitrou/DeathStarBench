// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

package xray

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// we can't know that the original driver will return driver.ErrSkip in advance.
// so we add this message to the query if it returns driver.ErrSkip.
const msgErrSkip = " -- skip fast-path; continue as if unimplemented"

// namedValueChecker is the same as driver.NamedValueChecker.
// Copied from database/sql/driver/driver.go for supporting Go 1.8.
type namedValueChecker interface {
	// CheckNamedValue is called before passing arguments to the driver
	// and is called in place of any ColumnConverter. CheckNamedValue must do type
	// validation and conversion as appropriate for the driver.
	CheckNamedValue(*driver.NamedValue) error
}

var (
	muInitializedDrivers sync.Mutex
	initializedDrivers   map[string]struct{}
	attrHook             func(attr *dbAttribute) // for testing
)

func initXRayDriver(driver, dsn string) error {
	muInitializedDrivers.Lock()
	defer muInitializedDrivers.Unlock()

	if initializedDrivers == nil {
		initializedDrivers = map[string]struct{}{}
	}
	if _, ok := initializedDrivers[driver]; ok {
		return nil
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	sql.Register(driver+":xray", &driverDriver{
		Driver:   db.Driver(),
		baseName: driver,
	})
	initializedDrivers[driver] = struct{}{}
	db.Close()
	return nil
}

// SQLContext opens a normalized and traced wrapper around an *sql.DB connection.
// It uses `sql.Open` internally and shares the same function signature.
// To ensure passwords are filtered, it is HIGHLY RECOMMENDED that your DSN
// follows the format: `<schema>://<user>:<password>@<host>:<port>/<database>`
func SQLContext(driver, dsn string) (*sql.DB, error) {
	if err := initXRayDriver(driver, dsn); err != nil {
		return nil, err
	}
	return sql.Open(driver+":xray", dsn)
}

type driverDriver struct {
	driver.Driver
	baseName string // the name of the base driver
}

func (d *driverDriver) Open(dsn string) (driver.Conn, error) {
	rawConn, err := d.Driver.Open(dsn)
	if err != nil {
		return nil, err
	}
	attr, err := newDBAttribute(context.Background(), d.baseName, d.Driver, rawConn, dsn, false)
	if err != nil {
		rawConn.Close()
		return nil, err
	}

	conn := &driverConn{
		Conn: rawConn,
		attr: attr,
	}
	return conn, nil
}

type driverConn struct {
	driver.Conn
	attr *dbAttribute
}

func (conn *driverConn) Ping(ctx context.Context) error {
	return Capture(ctx, conn.attr.dbname+conn.attr.host, func(ctx context.Context) error {
		conn.attr.populate(ctx, "PING")
		if p, ok := conn.Conn.(driver.Pinger); ok {
			return p.Ping(ctx)
		}
		return nil
	})
}

func (conn *driverConn) Prepare(query string) (driver.Stmt, error) {
	panic("not supported")
}

func (conn *driverConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	var stmt driver.Stmt
	var err error
	if connCtx, ok := conn.Conn.(driver.ConnPrepareContext); ok {
		stmt, err = connCtx.PrepareContext(ctx, query)
	} else {
		stmt, err = conn.Conn.Prepare(query)
		if err == nil {
			select {
			default:
			case <-ctx.Done():
				stmt.Close()
				return nil, ctx.Err()
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return &driverStmt{
		Stmt:  stmt,
		attr:  conn.attr,
		query: query,
		conn:  conn,
	}, nil
}

func (conn *driverConn) Begin() (driver.Tx, error) {
	panic("not supported")
}

func (conn *driverConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	var tx driver.Tx
	var err error
	if connCtx, ok := conn.Conn.(driver.ConnBeginTx); ok {
		tx, err = connCtx.BeginTx(ctx, opts)
	} else {
		if opts.Isolation != driver.IsolationLevel(sql.LevelDefault) {
			return nil, errors.New("xray: driver does not support non-default isolation level")
		}
		if opts.ReadOnly {
			return nil, errors.New("xray: driver does not support read-only transactions")
		}
		tx, err = conn.Conn.Begin()
		if err == nil {
			select {
			default:
			case <-ctx.Done():
				tx.Rollback()
				return nil, ctx.Err()
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return &driverTx{Tx: tx}, nil
}

func (conn *driverConn) Exec(query string, args []driver.Value) (driver.Result, error) {
	panic("not supported")
}

func (conn *driverConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	var err error
	var result driver.Result
	if execerCtx, ok := conn.Conn.(driver.ExecerContext); ok {
		Capture(ctx, conn.attr.dbname+conn.attr.host, func(ctx context.Context) error {
			result, err = execerCtx.ExecContext(ctx, query, args)
			if err == driver.ErrSkip {
				conn.attr.populate(ctx, query+msgErrSkip)
				return nil
			}
			conn.attr.populate(ctx, query)
			return err
		})
	} else {
		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		execer, ok := conn.Conn.(driver.Execer)
		if !ok {
			return nil, driver.ErrSkip
		}
		dargs, err0 := namedValuesToValues(args)
		if err0 != nil {
			return nil, err0
		}
		Capture(ctx, conn.attr.dbname+conn.attr.host, func(ctx context.Context) error {
			var err error
			result, err = execer.Exec(query, dargs)
			if err == driver.ErrSkip {
				conn.attr.populate(ctx, query+msgErrSkip)
				return nil
			}
			conn.attr.populate(ctx, query)
			return err
		})
	}
	return result, err
}

func (conn *driverConn) Query(query string, args []driver.Value) (driver.Rows, error) {
	panic("not supported")
}

func (conn *driverConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	var err error
	var rows driver.Rows
	if queryerCtx, ok := conn.Conn.(driver.QueryerContext); ok {
		Capture(ctx, conn.attr.dbname+conn.attr.host, func(ctx context.Context) error {
			rows, err = queryerCtx.QueryContext(ctx, query, args)
			if err == driver.ErrSkip {
				conn.attr.populate(ctx, query+msgErrSkip)
				return nil
			}
			conn.attr.populate(ctx, query)
			return err
		})
	} else {
		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		queryer, ok := conn.Conn.(driver.Queryer)
		if !ok {
			return nil, driver.ErrSkip
		}
		dargs, err0 := namedValuesToValues(args)
		if err0 != nil {
			return nil, err0
		}
		err = Capture(ctx, conn.attr.dbname+conn.attr.host, func(ctx context.Context) error {
			rows, err = queryer.Query(query, dargs)
			if err == driver.ErrSkip {
				conn.attr.populate(ctx, query+msgErrSkip)
				return nil
			}
			conn.attr.populate(ctx, query)
			return err
		})
	}
	return rows, err
}

func (conn *driverConn) Close() error {
	return conn.Conn.Close()
}

// copied from https://github.com/golang/go/blob/e6ebbe0d20fe877b111cf4ccf8349cba129d6d3a/src/database/sql/convert.go#L93-L99
// defaultCheckNamedValue wraps the default ColumnConverter to have the same
// function signature as the CheckNamedValue in the driver.NamedValueChecker
// interface.
func defaultCheckNamedValue(nv *driver.NamedValue) (err error) {
	nv.Value, err = driver.DefaultParameterConverter.ConvertValue(nv.Value)
	return err
}

// CheckNamedValue for implementing driver.NamedValueChecker
// This function may be unnecessary because `proxy.Stmt` already implements `NamedValueChecker`,
// but it is implemented just in case.
func (conn *driverConn) CheckNamedValue(nv *driver.NamedValue) (err error) {
	if nvc, ok := conn.Conn.(namedValueChecker); ok {
		return nvc.CheckNamedValue(nv)
	}
	// fallback to default
	return defaultCheckNamedValue(nv)
}

type dbAttribute struct {
	connectionString string
	url              string
	databaseType     string
	databaseVersion  string
	driverVersion    string
	user             string
	dbname           string
	host             string
}

func newDBAttribute(ctx context.Context, driverName string, d driver.Driver, conn driver.Conn, dsn string, filtered bool) (*dbAttribute, error) {
	var attr dbAttribute

	// Detect if DSN is a URL or not, set appropriate attribute
	urlDsn := dsn
	if !strings.Contains(dsn, "//") {
		urlDsn = "//" + urlDsn
	}
	// Here we're trying to detect things like `host:port/database` as a URL, which is pretty hard
	// So we just assume that if it's got a scheme, a user, or a query that it's probably a URL
	if u, err := url.Parse(urlDsn); err == nil && (u.Scheme != "" || u.User != nil || u.RawQuery != "" || strings.Contains(u.Path, "@")) {
		// Check that this isn't in the form of user/pass@host:port/db, as that will shove the host into the path
		if strings.Contains(u.Path, "@") {
			u, err = url.Parse(fmt.Sprintf("%s//%s%%2F%s", u.Scheme, u.Host, u.Path[1:]))
			if err != nil {
				return nil, err
			}
		}

		// Strip password from user:password pair in address
		if u.User != nil {
			uname := u.User.Username()

			// Some drivers use "user/pass@host:port" instead of "user:pass@host:port"
			// So we must manually attempt to chop off a potential password.
			// But we can skip this if we already found the password.
			if _, ok := u.User.Password(); !ok {
				uname = strings.Split(uname, "/")[0]
			}

			u.User = url.User(uname)
		}

		// Strip password and X-Amz-Security-Token (present in RDS IAM authentication) from query parameters
		q := u.Query()
		q.Del("password")
		q.Del("X-Amz-Security-Token")
		u.RawQuery = q.Encode()

		// In the case of known DSL sub segment name will be dbname@host
		host, _, _ := net.SplitHostPort(u.Host)
		if len(host) > 0 {
			attr.host = "@" + host
		} else {
			attr.host = host
		}

		attr.url = u.String()
		if !strings.Contains(dsn, "//") {
			attr.url = attr.url[2:]
		}
	} else {
		// We don't *think* it's a URL, so now we have to try our best to strip passwords from
		// some unknown DSL. We attempt to detect whether it's space-delimited or semicolon-delimited
		// then remove any keys with the name "password" or "pwd". This won't catch everything, but
		// from surveying the current (Jan 2017) landscape of drivers it should catch most.
		if filtered {
			attr.connectionString = dsn
		} else {
			attr.connectionString = stripPasswords(dsn)
		}
	}

	// Detect database type and use that to populate attributes
	var detectors []func(ctx context.Context, conn driver.Conn, attr *dbAttribute) error
	switch driverName {
	case "postgres":
		detectors = append(detectors, postgresDetector)
	case "mysql":
		detectors = append(detectors, mysqlDetector)
	default:
		detectors = append(detectors, postgresDetector, mysqlDetector, mssqlDetector, oracleDetector)
	}
	for _, detector := range detectors {
		if detector(ctx, conn, &attr) == nil {
			break
		}
		attr.databaseType = "Unknown"
		attr.databaseVersion = "Unknown"
		attr.user = "Unknown"
		attr.dbname = "Unknown"
	}

	// There's no standard to get SQL driver version information
	// So we invent an interface by which drivers can provide us this data
	type versionedDriver interface {
		Version() string
	}

	if vd, ok := d.(versionedDriver); ok {
		attr.driverVersion = vd.Version()
	} else {
		t := reflect.TypeOf(d)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		attr.driverVersion = t.PkgPath()
	}

	if attrHook != nil {
		attrHook(&attr)
	}
	return &attr, nil
}

func postgresDetector(ctx context.Context, conn driver.Conn, attr *dbAttribute) error {
	attr.databaseType = "Postgres"
	return queryRow(
		ctx, conn,
		"SELECT version(), current_user, current_database()",
		&attr.databaseVersion, &attr.user, &attr.dbname,
	)
}

func mysqlDetector(ctx context.Context, conn driver.Conn, attr *dbAttribute) error {
	attr.databaseType = "MySQL"
	return queryRow(
		ctx, conn,
		"SELECT version(), current_user(), database()",
		&attr.databaseVersion, &attr.user, &attr.dbname,
	)
}

func mssqlDetector(ctx context.Context, conn driver.Conn, attr *dbAttribute) error {
	attr.databaseType = "MS SQL"
	return queryRow(
		ctx, conn,
		"SELECT @@version, current_user, db_name()",
		&attr.databaseVersion, &attr.user, &attr.dbname,
	)
}

func oracleDetector(ctx context.Context, conn driver.Conn, attr *dbAttribute) error {
	attr.databaseType = "Oracle"
	return queryRow(
		ctx, conn,
		"SELECT version FROM v$instance UNION SELECT user, ora_database_name FROM dual",
		&attr.databaseVersion, &attr.user, &attr.dbname,
	)
}

// minimum implementation of (*sql.DB).QueryRow
func queryRow(ctx context.Context, conn driver.Conn, query string, dest ...*string) error {
	var err error

	// prepare
	var stmt driver.Stmt
	if connCtx, ok := conn.(driver.ConnPrepareContext); ok {
		stmt, err = connCtx.PrepareContext(ctx, query)
	} else {
		stmt, err = conn.Prepare(query)
		if err == nil {
			select {
			default:
			case <-ctx.Done():
				stmt.Close()
				return ctx.Err()
			}
		}
	}
	if err != nil {
		return err
	}
	defer stmt.Close()

	// execute query
	var rows driver.Rows
	if queryCtx, ok := stmt.(driver.StmtQueryContext); ok {
		rows, err = queryCtx.QueryContext(ctx, []driver.NamedValue{})
	} else {
		select {
		default:
		case <-ctx.Done():
			return ctx.Err()
		}
		rows, err = stmt.Query([]driver.Value{})
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	// scan
	if len(dest) != len(rows.Columns()) {
		return fmt.Errorf("xray: expected %d destination arguments in Scan, not %d", len(rows.Columns()), len(dest))
	}
	cols := make([]driver.Value, len(rows.Columns()))
	if err := rows.Next(cols); err != nil {
		return err
	}
	for i, src := range cols {
		d := dest[i]
		switch s := src.(type) {
		case string:
			*d = s
		case []byte:
			*d = string(s)
		case time.Time:
			*d = s.Format(time.RFC3339Nano)
		case int64:
			*d = strconv.FormatInt(s, 10)
		case float64:
			*d = strconv.FormatFloat(s, 'g', -1, 64)
		case bool:
			*d = strconv.FormatBool(s)
		default:
			return fmt.Errorf("sql: Scan error on column index %d, name %q: type missmatch", i, rows.Columns()[i])
		}
	}

	return nil
}

func (attr *dbAttribute) populate(ctx context.Context, query string) {
	seg := GetSegment(ctx)

	if seg == nil {
		processNilSegment(ctx)
		return
	}

	seg.Lock()
	seg.Namespace = "remote"
	seg.GetSQL().ConnectionString = attr.connectionString
	seg.GetSQL().URL = attr.url
	seg.GetSQL().DatabaseType = attr.databaseType
	seg.GetSQL().DatabaseVersion = attr.databaseVersion
	seg.GetSQL().DriverVersion = attr.driverVersion
	seg.GetSQL().User = attr.user
	seg.GetSQL().SanitizedQuery = query
	seg.Unlock()
}

type driverTx struct {
	driver.Tx
}

func (tx *driverTx) Commit() error {
	return tx.Tx.Commit()
}

func (tx *driverTx) Rollback() error {
	return tx.Tx.Rollback()
}

type driverStmt struct {
	driver.Stmt
	conn  *driverConn
	attr  *dbAttribute
	query string
}

func (stmt *driverStmt) Close() error {
	return stmt.Stmt.Close()
}

func (stmt *driverStmt) NumInput() int {
	return stmt.Stmt.NumInput()
}

func (stmt *driverStmt) Exec(args []driver.Value) (driver.Result, error) {
	panic("not supported")
}

func (stmt *driverStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	var result driver.Result
	var err error
	if execerContext, ok := stmt.Stmt.(driver.StmtExecContext); ok {
		err = Capture(ctx, stmt.attr.dbname+stmt.attr.host, func(ctx context.Context) error {
			stmt.populate(ctx)
			var err error
			result, err = execerContext.ExecContext(ctx, args)
			return err
		})
	} else {
		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		dargs, err0 := namedValuesToValues(args)
		if err0 != nil {
			return nil, err0
		}
		err = Capture(ctx, stmt.attr.dbname+stmt.attr.host, func(ctx context.Context) error {
			stmt.populate(ctx)
			var err error
			result, err = stmt.Stmt.Exec(dargs)
			return err
		})
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (stmt *driverStmt) Query(args []driver.Value) (driver.Rows, error) {
	panic("not supported")
}

func (stmt *driverStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	var result driver.Rows
	var err error
	if queryCtx, ok := stmt.Stmt.(driver.StmtQueryContext); ok {
		err = Capture(ctx, stmt.attr.dbname+stmt.attr.host, func(ctx context.Context) error {
			stmt.populate(ctx)
			var err error
			result, err = queryCtx.QueryContext(ctx, args)
			return err
		})
	} else {
		select {
		default:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		dargs, err0 := namedValuesToValues(args)
		if err0 != nil {
			return nil, err0
		}
		err = Capture(ctx, stmt.attr.dbname+stmt.attr.host, func(ctx context.Context) error {
			stmt.populate(ctx)
			var err error
			result, err = stmt.Stmt.Query(dargs)
			return err
		})
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (stmt *driverStmt) ColumnConverter(idx int) driver.ValueConverter {
	if conv, ok := stmt.Stmt.(driver.ColumnConverter); ok {
		return conv.ColumnConverter(idx)
	}
	return driver.DefaultParameterConverter
}

func (stmt *driverStmt) populate(ctx context.Context) {
	stmt.attr.populate(ctx, stmt.query)

	seg := GetSegment(ctx)

	if seg == nil {
		processNilSegment(ctx)
		return
	}

	seg.Lock()
	seg.GetSQL().Preparation = "statement"
	seg.Unlock()
}

// CheckNamedValue for implementing NamedValueChecker
func (stmt *driverStmt) CheckNamedValue(nv *driver.NamedValue) (err error) {
	if nvc, ok := stmt.Stmt.(namedValueChecker); ok {
		return nvc.CheckNamedValue(nv)
	}
	// When converting data in sql/driver/convert.go, it is checked first whether the `stmt`
	// implements `NamedValueChecker`, and then checks if `conn` implements NamedValueChecker.
	// In the case of "go-sql-proxy", the `proxy.Stmt` "implements" `CheckNamedValue` here,
	// so we also check both `stmt` and `conn` inside here.
	if nvc, ok := stmt.conn.Conn.(namedValueChecker); ok {
		return nvc.CheckNamedValue(nv)
	}
	// fallback to default
	return defaultCheckNamedValue(nv)
}

func namedValuesToValues(args []driver.NamedValue) ([]driver.Value, error) {
	var err error
	ret := make([]driver.Value, len(args))
	for _, arg := range args {
		if len(arg.Name) > 0 {
			err = errors.New("xray: driver does not support the use of Named Parameters")
		}
		ret[arg.Ordinal-1] = arg.Value
	}
	return ret, err
}

func stripPasswords(dsn string) string {
	var (
		tok        bytes.Buffer
		res        bytes.Buffer
		isPassword bool
		inBraces   bool
		delimiter  byte = ' '
	)
	flush := func() {
		if inBraces {
			return
		}
		if !isPassword {
			res.Write(tok.Bytes())
		}
		tok.Reset()
		isPassword = false
	}
	if strings.Count(dsn, ";") > strings.Count(dsn, " ") {
		delimiter = ';'
	}

	buf := strings.NewReader(dsn)
	for c, err := buf.ReadByte(); err == nil; c, err = buf.ReadByte() {
		tok.WriteByte(c)
		switch c {
		case ':', delimiter:
			flush()
		case '=':
			tokStr := strings.ToLower(tok.String())
			isPassword = `password=` == tokStr || `pwd=` == tokStr
			if b, err := buf.ReadByte(); err != nil {
				break
			} else {
				inBraces = b == '{'
			}
			if err := buf.UnreadByte(); err != nil {
				panic(err)
			}
		case '}':
			b, err := buf.ReadByte()
			if err != nil {
				break
			}
			if b == '}' {
				tok.WriteByte(b)
			} else {
				inBraces = false
				if err := buf.UnreadByte(); err != nil {
					panic(err)
				}
			}
		case '@':
			if strings.Contains(res.String(), ":") {
				resLen := res.Len()
				if resLen > 0 && res.Bytes()[resLen-1] == ':' {
					res.Truncate(resLen - 1)
				}
				isPassword = true
				flush()
				res.WriteByte(c)
			}
		}
	}
	inBraces = false
	flush()
	return res.String()
}

func processNilSegment(ctx context.Context) {
	cfg := GetRecorder(ctx)
	failedMessage := "failed to get segment from context since segment is nil"
	if cfg != nil && cfg.ContextMissingStrategy != nil {
		cfg.ContextMissingStrategy.ContextMissing(failedMessage)
	} else {
		globalCfg.ContextMissingStrategy().ContextMissing(failedMessage)
	}
}
