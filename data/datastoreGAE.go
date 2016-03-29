package data

import (
	"strings"
	"time"

	"google.golang.org/appengine/datastore"
)

// Store stores an item in the datastore
func (i *Item) Store(con *Context) error {
	i.Namespace = strings.ToLower(i.Namespace)
	k := datastore.NewKey(con.C, "Item", i.Namespace+i.URL, 0, nil)
	_, err := datastore.Put(con.C, k, i)
	if err != nil {
		con.Log.Errorf("Error while storing item in datastore. Item: %v. Error: %v", i, err)
		return err
	}
	i.DSKey = k.String()
	con.Log.Debugf("Stored item %+v", i)

	if err := clearCache(con, i.Namespace); err != nil {
		con.Log.Infof("Error clearing cache for namespace %v. Error: %v", i.Namespace, err)
	}

	return nil
}

// GetNewestItems returns the latest number elements for a specific namespace
func GetNewestItems(con *Context, namespace string, limit int, cursor string) ([]Item, string, error) {
	namespace = strings.ToLower(namespace)
	q := datastore.NewQuery("Item").
		Filter("Namespace =", namespace).
		Order("-CreatedAt").
		Limit(limit)

	return executeQuery(con, q, limit, cursor)
}

// GetNewestItemsByTime returns up to limit numbers of items stored >= the
// give time
func GetNewestItemsByTime(con *Context, namespace string, limit int, t time.Time, cursor string) ([]Item, string, error) {
	namespace = strings.ToLower(namespace)
	q := datastore.NewQuery("Item").
		Filter("Namespace =", namespace).
		Filter("CreatedAt >=", t).
		Order("-CreatedAt").
		Limit(limit)

	return executeQuery(con, q, limit, cursor)
}

func executeQuery(con *Context, q *datastore.Query, limit int, cursorStr string) ([]Item, string, error) {
	if cursor, err := datastore.DecodeCursor(cursorStr); err == nil {
		q = q.Start(cursor)
	}

	var is = make([]Item, 0, limit)
	var err error
	t := q.Run(con.C)
	for {
		var i Item
		_, err = t.Next(&i)
		if err == datastore.Done {
			break
		}

		is = append(is, i)
		if err != nil {
			con.Log.Errorf("Error fetching next item: %v", err)
			return nil, "", err
		}
	}

	var cursor datastore.Cursor
	if cursor, err = t.Cursor(); err == nil {
		return is, cursor.String(), nil
	}

	return nil, "", err
}

// NamespaceIsEmpty checks if there is already an item in a namespace
func NamespaceIsEmpty(con *Context, namespace string) (bool, error) {
	namespace = strings.ToLower(namespace)
	q := datastore.NewQuery("Item").
		Filter("Namespace =", namespace).
		Limit(1).
		KeysOnly()

	k, err := q.GetAll(con.C, nil)
	b := len(k) == 0

	return b, err
}

// ClearNamespace deletes every entry in a namespace
func ClearNamespace(con *Context, namespace string) error {
	namespace = strings.ToLower(namespace)
	q := datastore.NewQuery("Item").
		Filter("Namespace =", namespace).
		KeysOnly()

	k, err := q.GetAll(con.C, nil)
	if err != nil {
		return err
	}

	clearCache(con, namespace)

	con.Log.Infof("Going to delete %v items in the namespace %v", len(k), namespace)

	return datastore.DeleteMulti(con.C, k)
}

// DeleteAllItems deletes all items from datastore
func DeleteAllItems(con *Context) error {
	panic("Are you sure????!!!!")
	/*q := datastore.NewQuery("Item").KeysOnly()

	k, err := q.GetAll(con.C, nil)
	if err != nil {
		return err
	}

	return datastore.DeleteMulti(con.C, k)*/
}
