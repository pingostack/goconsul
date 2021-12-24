package goconsul

import (
	"fmt"
	"time"

	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/encoding/gtoml"
	"github.com/hashicorp/consul/api"
)

type KVCmd struct {
	client   *Client
	kv       *api.KV
	DataRoot string `json:"dataRoot"`
}

func NewKVCmd(client *Client, dataRoot string) *KVCmd {
	return &KVCmd{
		client:   client,
		kv:       client.KV(),
		DataRoot: dataRoot,
	}
}

func (cmd *KVCmd) realKey(key string) string {
	if cmd.DataRoot != "" {
		return fmt.Sprintf("%s/%s", cmd.DataRoot, key)
	}
	return key
}

func (cmd *KVCmd) GetStr(key string) (string, error) {
	if cmd.kv == nil {
		return "", fmt.Errorf("not found, center not connected")
	}

	key = cmd.realKey(key)
	keyPair, _, err := cmd.kv.Get(key, nil)
	if err != nil {
		return "", err
	}

	if keyPair == nil || keyPair.Value == nil {
		return "", fmt.Errorf("value of key[%s] is nil", key)
	}

	return string(keyPair.Value), nil
}

func (cmd *KVCmd) GetJson(key string, value interface{}) error {
	if cmd.kv == nil {
		return fmt.Errorf("not found, center not connected")
	}

	data, err := cmd.GetStr(key)
	if err != nil {
		return err
	}

	j := gjson.New(data)
	if j == nil {
		return fmt.Errorf("get config json[%s] failed, invalid json string", key)
	}

	err = j.Struct(value)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *KVCmd) GetToml(key string, value interface{}) error {
	if cmd.kv == nil {
		return fmt.Errorf("not found, center not connected")
	}

	data, err := cmd.GetStr(key)
	if err != nil {
		return err
	}

	err = gtoml.DecodeTo([]byte(data), value)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *KVCmd) SetStr(key string, value string) error {
	if cmd.kv == nil {
		return fmt.Errorf("not found, center not connected")
	}

	key = cmd.realKey(key)

	kv := &api.KVPair{
		Key:   key,
		Value: []byte(value),
	}

	_, err := cmd.kv.Put(kv, nil)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *KVCmd) SetJson(key string, value interface{}) error {
	if cmd.kv == nil {
		return fmt.Errorf("not found, center not connected")
	}

	data, err := gjson.Encode(value)
	if err != nil {
		return err
	}

	return cmd.SetStr(key, string(data))
}

func (cmd *KVCmd) SetToml(key string, value interface{}) error {
	if cmd.kv == nil {
		return fmt.Errorf("not found, center not connected")
	}

	data, err := gtoml.Encode(value)
	if err != nil {
		return err
	}

	return cmd.SetStr(key, string(data))
}

func (cmd *KVCmd) Delete(key string) error {
	realKey := fmt.Sprintf("%s%s", cmd.DataRoot, key)
	_, err := cmd.kv.Delete(realKey, nil)
	if err != nil {
		return err
	}

	return nil
}

func (cmd *KVCmd) Acquire(key string, value string, ttl string) (bool, error) {
	session := cmd.client.Session()
	sessionId, _, err := session.Create(&api.SessionEntry{
		Name:      key,
		Behavior:  api.SessionBehaviorDelete,
		TTL:       ttl,
		LockDelay: 1 * time.Microsecond,
	}, nil)

	if err != nil {
		return false, err
	}

	key = cmd.realKey(key)

	pair := &api.KVPair{
		Key:     key,
		Session: sessionId,
		Value:   ([]byte)(value),
	}

	blockCounter := 0

	for blockCounter < 2 {
		ok, _, err := cmd.kv.Acquire(pair, nil)

		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		}

		time.Sleep(time.Microsecond * 10)
		blockCounter++
	}

	return false, nil
}

func (cmd *KVCmd) UpdateExpire(key string) (bool, error) {
	if cmd.kv == nil {
		return false, fmt.Errorf("not found, center not connected")
	}

	realKey := cmd.realKey(key)
	keyPair, _, err := cmd.kv.Get(realKey, nil)
	if err != nil {
		return false, err
	}

	if keyPair == nil || keyPair.Value == nil {
		return false, fmt.Errorf("value of key[%s] is nil", realKey)
	}

	sessionId := keyPair.Session

	_, _, err = cmd.client.Session().Renew(sessionId, nil)
	if err != nil {
		return true, err
	}

	return true, nil
}
