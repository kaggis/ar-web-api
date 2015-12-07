/*
 * Copyright (c) 2015 GRNET S.A.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the
 * License. You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an "AS
 * IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 * express or implied. See the License for the specific language
 * governing permissions and limitations under the License.
 *
 * The views and conclusions contained in the software and
 * documentation are those of the authors and should not be
 * interpreted as representing official policies, either expressed
 * or implied, of either GRNET S.A., SRCE or IN2P3 CNRS Computing
 * Centre
 *
 * The work represented by this source file is partially funded by
 * the EGI-InSPIRE project through the European Commission's 7th
 * Framework Programme (contract # INFSO-RI-261323)
 */

package mongo

import (
	"errors"
	"log"
	"reflect"

	"github.com/ARGOeu/argo-web-api/Godeps/_workspace/src/github.com/twinj/uuid"
	"github.com/ARGOeu/argo-web-api/Godeps/_workspace/src/gopkg.in/mgo.v2"
	"github.com/ARGOeu/argo-web-api/Godeps/_workspace/src/gopkg.in/mgo.v2/bson"
)

type Id struct {
	ID bson.ObjectId `bson:"_id"`
}

func NewUUID() string {
	return uuid.NewV4().String()
}

func openCollection(session *mgo.Session, dbName string, collectionName string) *mgo.Collection {

	c := session.DB(dbName).C(collectionName)
	return c
}

func Pipe(session *mgo.Session, dbName string, collectionName string, query []bson.M, results interface{}) error {

	c := openCollection(session, dbName, collectionName)
	err := c.Pipe(query).All(results)
	return err
}

func Find(session *mgo.Session, dbName string, collectionName string, query interface{}, sorter string, results interface{}) error {
	c := openCollection(session, dbName, collectionName)
	var err error
	if sorter != "" {
		err = c.Find(query).Sort(sorter).All(results)
	} else {
		err = c.Find(query).All(results)
	}
	return err
}

func FindOne(session *mgo.Session, dbName string, collectionName string, query interface{}, result interface{}) error {
	c := openCollection(session, dbName, collectionName)
	var err error
	err = c.Find(query).One(result)

	return err
}

// FindAndProject uses a query and projection pair, updates results object and returns error code
func FindAndProject(session *mgo.Session, dbName string, collectionName string, query bson.M, projection bson.M, sorter string, results interface{}) error {

	c := openCollection(session, dbName, collectionName)
	err := c.Find(query).Select(projection).Sort(sorter).All(results)
	return err
}

//Insert a document into the collection of a database that are specified in the arguements
func Insert(session *mgo.Session, dbName string, collectionName string, query interface{}) error {

	c := openCollection(session, dbName, collectionName)
	err := c.Insert(query)
	return err
}

func MultiInsert(session *mgo.Session, dbName string, collectionName string, docs interface{}) error {
	array := structsToInterfaces(docs)
	c := openCollection(session, dbName, collectionName)
	err := c.Insert(array...)
	return err
}

func Remove(session *mgo.Session, dbName string, collectionName string, query bson.M) (*mgo.ChangeInfo, error) {

	c := openCollection(session, dbName, collectionName)
	info, err := c.RemoveAll(query)
	return info, err
}

func IdRemove(session *mgo.Session, dbName string, collectionName string, id string) error {
	// Check if given id is proper ObjectId
	if bson.IsObjectIdHex(id) == false {
		return errors.New("Invalid Object Id")
	}
	c := openCollection(session, dbName, collectionName)
	rid := bson.ObjectIdHex(id)
	err := c.RemoveId(rid)
	return err
}

func IdUpdate(session *mgo.Session, dbName string, collectionName string, id string, update interface{}) error {
	// Check if given id is proper ObjectId
	if bson.IsObjectIdHex(id) == false {
		return errors.New("Invalid Object Id")
	}
	c := openCollection(session, dbName, collectionName)
	rid := bson.ObjectIdHex(id)
	err := c.UpdateId(rid, update)
	return err
}

// Update a specfic document in a collection based on a query
func Update(session *mgo.Session, dbName string, collectionName string, query bson.M, update interface{}) error {
	// Check if given id is proper ObjectId
	c := openCollection(session, dbName, collectionName)
	err := c.Update(query, update)
	return err
}

// GetReportID accepts a report name and returns the report's id
func GetReportID(session *mgo.Session, dbName string, report string) (string, error) {
	var result map[string]interface{}
	// reports are stored to the reports collection
	c := openCollection(session, dbName, "reports")
	// query based on report name which is included in the info element
	query := bson.M{"info.name": report}
	// Execute the query and grab only the first result
	err := c.Find(query).One(&result)

	if result != nil {
		return result["id"].(string), err
	}
	return "", err

}

func structsToInterfaces(array interface{}) []interface{} {

	v := reflect.ValueOf(array)
	t := v.Type()

	if t.Kind() != reflect.Slice {
		log.Panicf("`array` should be %s but got %s", reflect.Slice, t.Kind())
	}

	result := make([]interface{}, v.Len(), v.Len())

	for i := 0; i < v.Len(); i++ {
		result[i] = v.Index(i).Interface()
	}

	return result
}
