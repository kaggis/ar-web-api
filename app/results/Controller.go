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
 * or implied, of GRNET S.A.
 *
 */

package results

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ARGOeu/argo-web-api/app/reports"
	"github.com/ARGOeu/argo-web-api/respond"
	"github.com/ARGOeu/argo-web-api/utils/config"
	"github.com/ARGOeu/argo-web-api/utils/mongo"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

// FlatListEndpointResults is responsible for handling request to flat list all available endpoint results
func FlatListEndpointResults(r *http.Request, cfg config.Config) (int, http.Header, []byte, error) {
	//STANDARD DECLARATIONS START
	code := http.StatusOK
	h := http.Header{}
	output := []byte("")
	err := error(nil)
	charset := "utf-8"
	//STANDARD DECLARATIONS END

	// Set Content-Type response Header value
	contentType := r.Header.Get("Accept")
	h.Set("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Parse the request into the input
	urlValues := r.URL.Query()
	vars := mux.Vars(r)
	supergroup := urlValues.Get("supergroup")
	service := urlValues.Get("service")

	skip := 0
	tkStr := urlValues.Get("nextPageToken")
	if tkStr != "" {
		tk, err := base64.StdEncoding.DecodeString(tkStr)
		if err != nil {
			code = http.StatusInternalServerError
			return code, h, output, err
		}
		skip, err = strconv.Atoi(string(tk))
		if err != nil {
			code = http.StatusInternalServerError
			return code, h, output, err
		}
	}

	limit := -1
	limStr := urlValues.Get("pageSize")
	if limStr != "" {
		limit, err = strconv.Atoi(limStr)
		if err != nil {
			code = http.StatusInternalServerError
			return code, h, output, err
		}
	}

	// Grab Tenant DB configuration from context
	tenantDbConfig := context.Get(r, "tenant_conf").(config.MongoConfig)

	session, err := mongo.OpenSession(tenantDbConfig)
	defer mongo.CloseSession(session)
	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	report := reports.MongoInterface{}
	err = mongo.FindOne(session, tenantDbConfig.Db, "reports", bson.M{"info.name": vars["report_name"]}, &report)

	if err != nil {
		code = http.StatusNotFound
		message := "The report with the name " + vars["report_name"] + " does not exist"
		output, err := createErrorMessage(message, code, contentType) //Render the response into XML or JSON
		return code, h, output, err
	}

	input := endpointResultQuery{
		basicQuery: basicQuery{
			Name: vars["endpoint_name"],

			Granularity: urlValues.Get("granularity"),
			Format:      contentType,
			StartTime:   urlValues.Get("start_time"),
			EndTime:     urlValues.Get("end_time"),
			Report:      report,
			Vars:        vars,
		},
		EndpointGroup: supergroup,
		Service:       service,
	}

	tenantDB := session.DB(tenantDbConfig.Db)
	errs := input.Validate(tenantDB)
	if len(errs) > 0 {
		out := respond.BadRequestSimple
		out.Errors = errs
		output = out.MarshalTo(contentType)
		code = 400
		return code, h, output, err
	}

	results := []EndpointInterface{}

	// Construct the query to mongodb based on the input
	filter := bson.M{
		"date":   bson.M{"$gte": input.StartTimeInt, "$lte": input.EndTimeInt},
		"report": report.ID,
	}

	if input.Name != "" {
		filter["name"] = input.Name
	}

	if input.Service != "" {
		filter["service"] = input.Service
	}

	if input.EndpointGroup != "" {
		filter["supergroup"] = input.EndpointGroup
	}

	// Select the granularity of the search daily/monthly
	custom := false
	if input.Granularity == "daily" {
		customForm[0] = "20060102"
		customForm[1] = "2006-01-02"
		query := FlatDailyEndpoint(filter, limit, skip)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_ar", query, &results)
	} else if input.Granularity == "monthly" {
		customForm[0] = "200601"
		customForm[1] = "2006-01"
		query := FlatMonthlyEndpoint(filter, limit, skip)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_ar", query, &results)

	} else if input.Granularity == "custom" {
		customForm[0] = "20060102"
		customForm[1] = "2006-01-02"
		query := FlatCustomEndpoint(filter, limit, skip)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_ar", query, &results)
		custom = true
	}

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	if len(results) == 0 {
		code = http.StatusNotFound
		message := "No results found for given query"
		output, err = createErrorMessage(message, code, contentType)
		return code, h, output, err
	}

	output, err = createFlatEndpointResultView(results, report, input.Format, limit, skip, custom)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	return code, h, output, err

}

// ListEndpointResults is responsible for handling request to list service flavor results
func ListEndpointResults(r *http.Request, cfg config.Config) (int, http.Header, []byte, error) {
	//STANDARD DECLARATIONS START
	code := http.StatusOK
	h := http.Header{}
	output := []byte("")
	err := error(nil)
	charset := "utf-8"
	//STANDARD DECLARATIONS END

	// Set Content-Type response Header value
	contentType := r.Header.Get("Accept")
	h.Set("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Parse the request into the input
	urlValues := r.URL.Query()
	vars := mux.Vars(r)

	// Grab Tenant DB configuration from context
	tenantDbConfig := context.Get(r, "tenant_conf").(config.MongoConfig)

	session, err := mongo.OpenSession(tenantDbConfig)
	defer mongo.CloseSession(session)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	report := reports.MongoInterface{}
	err = mongo.FindOne(session, tenantDbConfig.Db, "reports", bson.M{"info.name": vars["report_name"]}, &report)

	if err != nil {
		code = http.StatusNotFound
		message := "The report with the name " + vars["report_name"] + " does not exist"
		output, err := createErrorMessage(message, code, contentType) //Render the response into XML or JSON
		return code, h, output, err
	}

	input := endpointResultQuery{
		basicQuery: basicQuery{
			Name: vars["endpoint_name"],

			Granularity: urlValues.Get("granularity"),
			Format:      contentType,
			StartTime:   urlValues.Get("start_time"),
			EndTime:     urlValues.Get("end_time"),
			Report:      report,
			Vars:        vars,
		},
		EndpointGroup: vars["lgroup_name"],
		Service:       vars["service_type"],
	}

	tenantDB := session.DB(tenantDbConfig.Db)
	errs := input.Validate(tenantDB)
	if len(errs) > 0 {
		out := respond.BadRequestSimple
		out.Errors = errs
		output = out.MarshalTo(contentType)
		code = 400
		return code, h, output, err
	}

	if vars["lgroup_type"] != report.GetEndpointGroupType() {
		code = http.StatusNotFound
		message := "The report " + vars["report_name"] + " does not define endpoint group type: " + vars["lgroup_type"] + ". Try using " + report.GetEndpointGroupType() + " instead."
		output, err := createErrorMessage(message, code, contentType) //Render the response into XML or JSON
		return code, h, output, err
	}

	results := []EndpointInterface{}

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	// Construct the query to mongodb based on the input
	filter := bson.M{
		"date":   bson.M{"$gte": input.StartTimeInt, "$lte": input.EndTimeInt},
		"report": report.ID,
	}

	if input.Name != "" {
		filter["name"] = input.Name
	}

	if input.EndpointGroup != "" {
		filter["supergroup"] = input.EndpointGroup
	}

	if input.Service != "" {
		filter["service"] = input.Service
	}

	// Select the granularity of the search daily/monthly
	custom := false
	if input.Granularity == "daily" {
		customForm[0] = "20060102"
		customForm[1] = "2006-01-02"
		query := DailyEndpoint(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_ar", query, &results)
	} else if input.Granularity == "monthly" {
		customForm[0] = "200601"
		customForm[1] = "2006-01"
		query := MonthlyEndpoint(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_ar", query, &results)
	} else if input.Granularity == "custom" {
		customForm[0] = "200601"
		customForm[1] = "2006-01"
		query := CustomEndpoint(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_ar", query, &results)
		custom = true
	}

	// mongo.Find(session, tenantDbConfig.Db, "endpoint_group_ar", bson.M{}, "_id", &results)
	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	if len(results) == 0 {
		code = http.StatusNotFound
		message := "No results found for given query"
		output, err = createErrorMessage(message, code, contentType)
		return code, h, output, err
	}

	output, err = createEndpointResultView(results, report, input.Format, custom)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	return code, h, output, err

}

// ListServiceFlavorResults is responsible for handling request to list service flavor results
func ListServiceFlavorResults(r *http.Request, cfg config.Config) (int, http.Header, []byte, error) {
	//STANDARD DECLARATIONS START
	code := http.StatusOK
	h := http.Header{}
	output := []byte("")
	err := error(nil)
	charset := "utf-8"
	//STANDARD DECLARATIONS END

	// Set Content-Type response Header value
	contentType := r.Header.Get("Accept")
	h.Set("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Parse the request into the input
	urlValues := r.URL.Query()
	vars := mux.Vars(r)

	// Grab Tenant DB configuration from context
	tenantDbConfig := context.Get(r, "tenant_conf").(config.MongoConfig)

	session, err := mongo.OpenSession(tenantDbConfig)
	defer mongo.CloseSession(session)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	report := reports.MongoInterface{}
	err = mongo.FindOne(session, tenantDbConfig.Db, "reports", bson.M{"info.name": vars["report_name"]}, &report)

	if err != nil {
		code = http.StatusNotFound
		message := "The report with the name " + vars["report_name"] + " does not exist"
		output, err := createErrorMessage(message, code, contentType) //Render the response into XML or JSON
		return code, h, output, err
	}

	input := serviceFlavorResultQuery{
		basicQuery: basicQuery{
			Name:        vars["service_type"],
			Granularity: urlValues.Get("granularity"),
			Format:      contentType,
			StartTime:   urlValues.Get("start_time"),
			EndTime:     urlValues.Get("end_time"),
			Report:      report,
			Vars:        vars,
		},
		EndpointGroup: vars["lgroup_name"],
	}

	tenantDB := session.DB(tenantDbConfig.Db)
	errs := input.Validate(tenantDB)
	if len(errs) > 0 {
		out := respond.BadRequestSimple
		out.Errors = errs
		output = out.MarshalTo(contentType)
		code = 400
		return code, h, output, err
	}

	if vars["lgroup_type"] != report.GetEndpointGroupType() {
		code = http.StatusNotFound
		message := "The report " + vars["report_name"] + " does not define endpoint group type: " + vars["lgroup_type"] + ". Try using " + report.GetEndpointGroupType() + " instead."
		output, err := createErrorMessage(message, code, contentType) //Render the response into XML or JSON
		return code, h, output, err
	}

	results := []ServiceFlavorInterface{}

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	// Construct the query to mongodb based on the input
	filter := bson.M{
		"date":   bson.M{"$gte": input.StartTimeInt, "$lte": input.EndTimeInt},
		"report": report.ID,
	}

	if input.Name != "" {
		filter["name"] = input.Name
	}

	if input.EndpointGroup != "" {
		filter["supergroup"] = input.EndpointGroup
	}

	// Select the granularity of the search daily/monthly
	custom := false
	if input.Granularity == "daily" {
		customForm[0] = "20060102"
		customForm[1] = "2006-01-02"
		query := DailyServiceFlavor(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "service_ar", query, &results)
	} else if input.Granularity == "monthly" {
		customForm[0] = "200601"
		customForm[1] = "2006-01"
		query := MonthlyServiceFlavor(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "service_ar", query, &results)
	} else if input.Granularity == "custom" {
		customForm[0] = "200601"
		customForm[1] = "2006-01"
		query := CustomServiceFlavor(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "service_ar", query, &results)
		custom = true
	}

	// mongo.Find(session, tenantDbConfig.Db, "endpoint_group_ar", bson.M{}, "_id", &results)
	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	if len(results) == 0 {
		code = http.StatusNotFound
		message := "No results found for given query"
		output, err = createErrorMessage(message, code, contentType)
		return code, h, output, err
	}

	output, err = createServiceFlavorResultView(results, report, input.Format, custom)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	return code, h, output, err

}

// ListEndpointGroupResults endpoint group availabilities according to the http request
func ListEndpointGroupResults(r *http.Request, cfg config.Config) (int, http.Header, []byte, error) {

	//STANDARD DECLARATIONS START
	code := http.StatusOK
	h := http.Header{}
	output := []byte("")
	err := error(nil)
	charset := "utf-8"
	//STANDARD DECLARATIONS END

	// Set Content-Type response Header value
	contentType := r.Header.Get("Accept")
	h.Set("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Parse the request into the input
	urlValues := r.URL.Query()
	vars := mux.Vars(r)

	// Grab Tenant DB configuration from context
	tenantDbConfig := context.Get(r, "tenant_conf").(config.MongoConfig)

	session, err := mongo.OpenSession(tenantDbConfig)
	defer mongo.CloseSession(session)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	report := reports.MongoInterface{}
	err = mongo.FindOne(session, tenantDbConfig.Db, "reports", bson.M{"info.name": vars["report_name"]}, &report)

	if err != nil {
		code = http.StatusNotFound
		message := "The report with the name " + vars["report_name"] + " does not exist"
		output, err := createErrorMessage(message, code, contentType) //Render the response into XML or JSON
		h.Set("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))
		return code, h, output, err
	}

	input := endpointGroupResultQuery{
		basicQuery{
			Name:        vars["lgroup_name"],
			Granularity: urlValues.Get("granularity"),
			Format:      contentType,
			StartTime:   urlValues.Get("start_time"),
			EndTime:     urlValues.Get("end_time"),
			Report:      report,
			Vars:        vars,
		}, "",
	}

	tenantDB := session.DB(tenantDbConfig.Db)
	errs := input.Validate(tenantDB)
	if len(errs) > 0 {
		out := respond.BadRequestSimple
		out.Errors = errs
		output = out.MarshalTo(contentType)
		code = 400
		return code, h, output, err
	}

	if vars["lgroup_type"] != report.GetEndpointGroupType() {
		code = http.StatusNotFound
		message := "The report " + vars["report_name"] + " does not define endpoint group type: " + vars["lgroup_type"] + ". Try using " + report.GetEndpointGroupType() + " instead."
		output, err := createErrorMessage(message, code, contentType) //Render the response into XML or JSON
		return code, h, output, err
	}

	results := []EndpointGroupInterface{}

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	// Construct the query to mongodb based on the input
	filter := bson.M{
		"date":   bson.M{"$gte": input.StartTimeInt, "$lte": input.EndTimeInt},
		"report": report.ID,
	}

	if input.Name != "" {
		filter["name"] = input.Name
	}

	// Select the granularity of the search daily/monthly
	custom := false
	if input.Granularity == "daily" {
		customForm[0] = "20060102"
		customForm[1] = "2006-01-02"
		query := DailyEndpointGroup(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_group_ar", query, &results)
	} else if input.Granularity == "monthly" {
		customForm[0] = "200601"
		customForm[1] = "2006-01"
		query := MonthlyEndpointGroup(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_group_ar", query, &results)
	} else if input.Granularity == "custom" {
		customForm[0] = "200601"
		customForm[1] = "2006-01"
		query := CustomEndpointGroup(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_group_ar", query, &results)
		custom = true
	}

	// mongo.Find(session, tenantDbConfig.Db, "endpoint_group_ar", bson.M{}, "_id", &results)
	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	if len(results) == 0 {
		code = http.StatusNotFound
		message := "No results found for given query"
		output, err = createErrorMessage(message, code, contentType)
		return code, h, output, err
	}

	output, err = createEndpointGroupResultView(results, report, input.Format, custom)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	return code, h, output, err
}

// ListSuperGroupResults supergroup availabilities according to the http request
func ListSuperGroupResults(r *http.Request, cfg config.Config) (int, http.Header, []byte, error) {

	//STANDARD DECLARATIONS START
	code := http.StatusOK
	h := http.Header{}
	output := []byte("")
	err := error(nil)
	charset := "utf-8"
	//STANDARD DECLARATIONS END

	// Set Content-Type response Header value
	contentType := r.Header.Get("Accept")
	h.Set("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))

	// Parse the request into the input
	urlValues := r.URL.Query()
	vars := mux.Vars(r)

	// Grab Tenant DB configuration from context
	tenantDbConfig := context.Get(r, "tenant_conf").(config.MongoConfig)

	session, err := mongo.OpenSession(tenantDbConfig)
	defer mongo.CloseSession(session)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	report := reports.MongoInterface{}
	err = mongo.FindOne(session, tenantDbConfig.Db, "reports", bson.M{"info.name": vars["report_name"]}, &report)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}
	input := endpointGroupResultQuery{
		basicQuery{
			Name:        vars["group_name"],
			Granularity: urlValues.Get("granularity"),
			Format:      contentType,
			StartTime:   urlValues.Get("start_time"),
			EndTime:     urlValues.Get("end_time"),
			Report:      report,
			Vars:        vars,
		}, "",
	}

	tenantDB := session.DB(tenantDbConfig.Db)
	errs := input.Validate(tenantDB)
	if len(errs) > 0 {
		out := respond.BadRequestSimple
		out.Errors = errs
		output = out.MarshalTo(contentType)
		code = 400
		return code, h, output, err
	}

	results := []SuperGroupInterface{}

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	// Construct the query to mongodb based on the input
	filter := bson.M{
		"date":   bson.M{"$gte": input.StartTimeInt, "$lte": input.EndTimeInt},
		"report": report.ID,
	}

	if input.Name != "" {
		filter["supergroup"] = input.Name
	}

	custom := false
	// Select the granularity of the search daily/monthly
	if input.Granularity == "daily" {
		customForm[0] = "20060102"
		customForm[1] = "2006-01-02"
		query := DailySuperGroup(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_group_ar", query, &results)
	} else if input.Granularity == "monthly" {
		customForm[0] = "200601"
		customForm[1] = "2006-01"
		query := MonthlySuperGroup(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_group_ar", query, &results)
	} else if input.Granularity == "custom" {
		customForm[0] = "200601"
		customForm[1] = "2006-01"
		query := CustomSuperGroup(filter)
		err = mongo.Pipe(session, tenantDbConfig.Db, "endpoint_group_ar", query, &results)
		custom = true
	}
	// mongo.Find(session, tenantDbConfig.Db, "endpoint_group_ar", bson.M{}, "_id", &results)
	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	if len(results) == 0 {
		code = http.StatusNotFound
		message := "No results found for given query"
		output, err = createErrorMessage(message, code, contentType)
		return code, h, output, err
	}

	output, err = createSuperGroupView(results, report, input.Format, custom)

	if err != nil {
		code = http.StatusInternalServerError
		return code, h, output, err
	}

	return code, h, output, err
}

func Options(r *http.Request, cfg config.Config) (int, http.Header, []byte, error) {

	//STANDARD DECLARATIONS START

	code := http.StatusOK
	h := http.Header{}
	output := []byte("")
	err := error(nil)
	contentType := "text/plain"
	charset := "utf-8"

	//STANDARD DECLARATIONS END

	h.Set("Content-Type", fmt.Sprintf("%s; charset=%s", contentType, charset))
	h.Set("Allow", fmt.Sprintf("GET, OPTIONS"))
	return code, h, output, err

}
