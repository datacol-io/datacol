package client

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dinesh/datacol/client/models"
	"github.com/dinesh/datacol/cloud"
	"github.com/dinesh/datacol/cmd/stdcli"
	"github.com/joyrexus/buckets"
)


func init() {
	root := models.ConfigPath
	if _, err := os.Stat(root); err != nil {
		if !os.IsNotExist(err) {
			stdcli.Error(err)
			return
		} else {
			if err := os.MkdirAll(root, 0700); err != nil {
				stdcli.Error(err)
				return
			}
		}
	}
}

func DBkv() *buckets.DB {
	dbpath := filepath.Join(models.ConfigPath, models.DbFilename)
	db, err := buckets.Open(dbpath)
	if err != nil {
		stdcli.Error(fmt.Errorf("creating database file: %v", err))
		return nil
	}

	return db
}

func getV(bk, key []byte) ([]byte, error) {
	store := DBkv()
	defer store.Close()

	getter, err := store.New(bk)
	if err != nil {
		return nil, err
	}

	return getter.Get(key)
}

func deleteV(bk []byte, key string) error {
	store := DBkv()
	defer store.Close()

	bx, _ := store.New(bk)
	return bx.Delete([]byte(key))
}

func getList(bk []byte) ([]buckets.Item, error) {
	store := DBkv()
	defer store.Close()

	abx, _ := store.New(bk)
	return abx.Items()
}

func Persist(b []byte, pk string, object interface{}) error {
	store := DBkv()
	defer store.Close()

	bx, _ := store.New(b)
	encoded, err := json.Marshal(object)
	if err != nil {
		return err
	}

	return bx.Put([]byte(pk), encoded)
}

type Client struct {
	Version   string
	StackName string
	*Stack
}

func (c *Client) configRoot() string {
	return filepath.Join(models.ConfigPath, c.StackName)
}

func (c *Client) SetStack(name string) error {
	c.StackName = name
	st, err := FindStack(name)
	if err != nil {
		return stdcli.Stack404
	}

	c.Stack = st
	return nil
}

func (c *Client) Provider() cloud.Provider {
	if c.Stack == nil {
		log.Fatal(stdcli.Stack404)
	}

	return cloud.Getgcp(
		c.Stack.Name,
		c.Stack.ProjectId,
		c.Stack.PNumber,
		c.Stack.Zone,
		c.Stack.Bucket,
		c.Stack.ServiceKey,
	)
}
