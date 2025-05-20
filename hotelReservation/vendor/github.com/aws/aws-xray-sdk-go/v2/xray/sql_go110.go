// Copyright 2017-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may not use this file except in compliance with the License. A copy of the License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

// +build go1.10

package xray

import (
	"context"
	"database/sql/driver"
	"sync"
)

// SQLConnector wraps the connector, and traces SQL executions.
// Unlike SQLContext, SQLConnector doesn't filter the password of the dsn.
// So, you have to filter the password before passing the dsn to SQLConnector.
func SQLConnector(dsn string, connector driver.Connector) driver.Connector {
	d := &driverDriver{
		Driver:   connector.Driver(),
		baseName: "unknown",
	}
	return &driverConnector{
		Connector: connector,
		driver:    d,
		name:      dsn,
		filtered:  true,
		// initialized attr lazy because we have no context here.
	}
}

func (conn *driverConn) ResetSession(ctx context.Context) error {
	if sr, ok := conn.Conn.(driver.SessionResetter); ok {
		return sr.ResetSession(ctx)
	}
	return nil
}

type driverConnector struct {
	driver.Connector
	driver   *driverDriver
	filtered bool
	name     string

	mu   sync.RWMutex
	attr *dbAttribute
}

func (c *driverConnector) Connect(ctx context.Context) (driver.Conn, error) {
	var rawConn driver.Conn
	attr, err := c.getAttr(ctx)
	if err != nil {
		return nil, err
	}
	err = Capture(ctx, attr.dbname+attr.host, func(ctx context.Context) error {
		attr.populate(ctx, "CONNECT")
		var err error
		rawConn, err = c.Connector.Connect(ctx)
		return err
	})
	if err != nil {
		return nil, err
	}

	conn := &driverConn{
		Conn: rawConn,
		attr: attr,
	}
	return conn, nil
}

func (c *driverConnector) getAttr(ctx context.Context) (*dbAttribute, error) {
	c.mu.RLock()
	attr := c.attr
	c.mu.RUnlock()
	if attr != nil {
		return attr, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.attr != nil {
		return c.attr, nil
	}
	conn, err := c.Connector.Connect(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	attr, err = newDBAttribute(ctx, c.driver.baseName, c.driver.Driver, conn, c.name, c.filtered)
	if err != nil {
		return nil, err
	}
	c.attr = attr
	return attr, nil
}

func (c *driverConnector) Driver() driver.Driver {
	return c.driver
}

type fallbackConnector struct {
	driver driver.Driver
	name   string
}

func (c *fallbackConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.driver.Open(c.name)
	if err != nil {
		return nil, err
	}
	select {
	default:
	case <-ctx.Done():
		conn.Close()
		return nil, ctx.Err()
	}
	return conn, nil
}

func (c *fallbackConnector) Driver() driver.Driver {
	return c.driver
}

func (d *driverDriver) OpenConnector(name string) (driver.Connector, error) {
	var c driver.Connector
	if dctx, ok := d.Driver.(driver.DriverContext); ok {
		var err error
		c, err = dctx.OpenConnector(name)
		if err != nil {
			return nil, err
		}
	} else {
		c = &fallbackConnector{
			driver: d.Driver,
			name:   name,
		}
	}
	c = &driverConnector{
		Connector: c,
		driver:    d,
		filtered:  false,
		name:      name,
		// initialized attr lazy because we have no context here.
	}
	return c, nil
}
