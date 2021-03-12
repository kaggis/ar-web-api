/*
 * Copyright (c) 2014 GRNET S.A., SRCE, IN2P3 CNRS Computing Centre
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

package tenants

// Tenant structure holds information about tenant information
// including db conf and users. Used in
type Tenant struct {
	ID       string         `bson:"id" json:"id"`
	Info     TenantInfo     `bson:"info" json:"info"`
	DbConf   []TenantDbConf `bson:"db_conf" json:"db_conf,omitempty"`
	Topology TopologyInfo   `bson:"topology" json:"topology"`
	Users    []TenantUser   `bson:"users" json:"users,omitempty"`
}

// TopologyInfo contains topology feed information
type TopologyInfo struct {
	TopoType string `bson:"type" json:"type"`
	Feed     string `bson:"feed" json:"feed"`
}

type TenantStatus struct {
	ID     string       `bson:"id" json:"id"`
	Info   TenantInfo   `bson:"info" json:"info"`
	Status StatusDetail `bson:"status" json:"status,omitempty"`
}

type StatusDetail struct {
	TotalStatus  bool        `json:"total_status"`
	AMS          DetailsAMS  `bson:"ams" json:"ams"`
	HDFS         DetailsHDFS `bson:"hdfs" json:"hdfs"`
	API          DetailAPI   `bson:"api" json:"api"`
	EngineConfig bool        `bson:"engine_config" json:"engine_config"`
	LastCheck    string      `bson:"last_check" json:"last_check"`
}

type DetailsAMS struct {
	MetricData NodeAMS `bson:"metric_data" json:"metric_data"`
}

type DetailsHDFS struct {
	MetricData bool `bson:"metric_data" json:"metric_data"`
}

type DetailAPI struct {
	AggregationProf bool `bson:"aggregation_profiles" json:"aggregation_profiles"`
	ConfigProf      bool `bson:"reports" json:"reports"`
	Donwtimes       bool `bson:"downtimes" json:"downtimes"`
	GroupEndpoints  bool `bson:"group_endpoints" json:"group_endpoints"`
	GroupGroups     bool `bson:"group_groups" json:"group_groups"`
	MetricProf      bool `bson:"metric_profiles" json:"metric_profiles"`
	OpsProf         bool `bson:"operations_profiles" json:"operations_profiles"`
	Weight          bool `bson:"weights" json:"weights"`
}

type NodeAMS struct {
	Ingestion       bool  `bson:"ingestion" json:"ingestion"`
	Publishing      bool  `bson:"publishing" json:"publishing"`
	StatusStreaming bool  `bson:"status_streaming" json:"status_streaming"`
	MsgArrived      int64 `bson:"messages_arrived" json:"messages_arrived"`
}

// TenantInfo struct holds information about tenant name, contact details
type TenantInfo struct {
	Name        string `bson:"name" json:"name"`
	Email       string `bson:"email" json:"email"`
	Description string `bson:"description" json:"description"`
	Image       string `bson:"image" json:"image"`
	Website     string `bson:"website" json:"website"`
	Created     string `bson:"created" json:"created"`
	Updated     string `bson:"updated" json:"updated"`
}

// TenantDbConf structure holds information about tenant's
// database configuration
type TenantDbConf struct {
	Store    string `bson:"store" json:"store"`
	Server   string `bson:"server" json:"server"`
	Port     int    `bson:"port" json:"port"`
	Database string `bson:"database" json:"database"`
	Username string `bson:"username" json:"username"`
	Password string `bson:"password" json:"password"`
}

// TenantUser structure holds information about tenant's
// user
type TenantUser struct {
	ID     string   `bson:"id" json:"id"`
	Name   string   `bson:"name"       json:"name"`
	Email  string   `bson:"email"      json:"email"`
	APIkey string   `bson:"api_key"    json:"api_key"`
	Roles  []string `bson:"roles,omitempty"      json:"roles,omitempty"`
}

// SelfReference to hold links and id
type SelfReference struct {
	ID    string `json:"id" bson:"id,omitempty"`
	Links Links  `json:"links"`
}

// Links struct to hold links
type Links struct {
	Self string `json:"self"`
}

// Token

type Token struct {
	APIkey string `json:"api_key"`
}
