package etcd

import (
	"context"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/juju/errors"
	"github.com/zssky/log"
)

const (
	// etcdKeyTTL 1秒超时
	etcdKeyTTL = 1
)

// Client etcd client.
type Client struct {
	client *clientv3.Client
}

var (
	//networkTimeout 超时.
	networkTimeout = time.Second * 3
)

// New new etcd client.
func New(addrs []string) (*Client, error) {
	c, err := clientv3.New(clientv3.Config{
		Endpoints:   addrs,
		DialTimeout: networkTimeout,
	})
	if err != nil {
		return nil, errors.Annotatef(err, "addrs:%+v", addrs)
	}

	return &Client{client: c}, nil
}

// Get get value from etcd.
func (e *Client) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), networkTimeout)
	resp, err := clientv3.NewKV(e.client).Get(ctx, key)
	cancel()
	if err != nil {
		return "", errors.Trace(err)
	}

	if len(resp.Kvs) == 0 {
		log.Debugf("key:%s value not found", key)
		return "0", nil
	}

	log.Debugf("find key:%s, value:%s", key, string(resp.Kvs[0].Value))
	return string(resp.Kvs[0].Value), nil
}

// List get keys from etcd.
func (e *Client) List(prefix string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), networkTimeout)
	resp, err := clientv3.NewKV(e.client).Get(ctx, prefix, clientv3.WithPrefix())
	cancel()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if len(resp.Kvs) == 0 {
		log.Debugf("prefix:%s value not found", prefix)
		return nil, errors.New("not found")
	}

	keys := make(map[string]string)
	for _, k := range resp.Kvs {
		keys[string(k.Key)] = string(k.Value)
		log.Debugf("find key:%s", k.Key)
	}

	return keys, nil
}

// CAS put value to etcd.
func (e *Client) CAS(cmpKey, cmpValue, key, value string) error {
	cmp := clientv3.Compare(clientv3.Value(cmpKey), "=", cmpValue)
	if cmpValue == "" {
		cmp = clientv3.Compare(clientv3.CreateRevision(cmpKey), "=", 0)
	}
	ctx, cancel := context.WithTimeout(context.Background(), networkTimeout)
	pr, err := e.client.Txn(ctx).
		If(cmp).
		Then(clientv3.OpPut(key, value)).
		Commit()
	cancel()
	if err != nil {
		return errors.Trace(err)
	}

	if !pr.Succeeded {
		return errors.New("put key failed")
	}

	return nil
}

//WatchPrefix 监控指定前缀.
func (e *Client) WatchPrefix(key string, ec chan clientv3.Event) {
	watcher := clientv3.NewWatcher(e.client)
	defer watcher.Close()

	for resp := range watcher.Watch(e.client.Ctx(), key, clientv3.WithPrefix()) {
		if resp.Canceled {
			return
		}
		log.Debugf("resp:%+v", resp)
		for _, e := range resp.Events {
			ec <- *e
		}
	}
}

//Put 写.
func (e *Client) Put(key, val string) error {
	ctx, cancel := context.WithTimeout(context.Background(), networkTimeout)
	pr, err := e.client.Put(ctx, key, val)
	cancel()
	if err != nil {
		return errors.Trace(err)
	}

	log.Debugf("pr:%v", pr)

	return nil
}

//Keepalive 创建并保活一个key.
func (e *Client) Keepalive(key, val string) (clientv3.Lease, error) {
	lessor := clientv3.NewLease(e.client)

	ctx, cancel := context.WithTimeout(context.Background(), networkTimeout)
	lr, err := lessor.Grant(ctx, etcdKeyTTL)
	cancel()
	if err != nil {
		lessor.Close()
		return nil, errors.Trace(err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), networkTimeout)
	pr, err := e.client.Put(ctx, key, val, clientv3.WithLease(clientv3.LeaseID(lr.ID)))
	cancel()
	if err != nil {
		lessor.Close()
		return nil, errors.Trace(err)
	}

	log.Debugf("pr:%v", pr)

	ctx, cancel = context.WithCancel(context.Background())
	if _, err = lessor.KeepAlive(ctx, clientv3.LeaseID(lr.ID)); err != nil {
		lessor.Close()
		return nil, errors.Trace(err)
	}

	log.Debugf("keepalive:%v, key:%v", lr.ID, key)

	return lessor, nil

}

// Close 关闭客户端
func (e *Client) Close() {
	e.client.Close()
}
