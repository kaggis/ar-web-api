---
id: v3_status_results
title: Status Results (v3)
---

## API Calls

_Note_: These are v3 api calls implementations found under the path `/api/v3`

| Name                                                                          | Description                                                                                                                                                                                                                              | Shortcut          |
| ----------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------- |
| GET: List Status Results | This method retrieves the status results of all top level supergroups and their included endpoints | [Description](#1) |

<a id="1"></a>

# [GET]: List Status results for top level groupds and inclded endpoints

The following methods can be used to obtain a tenant's Status timeline results for all top level groups and included endpoints. The api authenticates the tenant using the api-key within the x-api-key header. User can specify time granularity (`monthly` or `daily`) for retrieved results and also format using the `Accept` header. 

### Input

```
/status/{report_name}?start_time={value}&end_time={value}&view={value}
```

#### Query Parameters

| Type            | Description                                                                                     | Required | Default value |
| --------------- | ----------------------------------------------------------------------------------------------- | -------- | ------------- |
| `start_time`  | UTC time in W3C format                                                                          | NO       | beginning of the current day (UTC, W3C format)
| `end_time`    |  UTC time in W3C format                                                                          | NO       | time right now (UTC, W3C format)
| `view`        | `view=details` to display verbose details in results (such as threshold rules etc.) and `view=latest` to display only display the latest status value for each item                                                                        | NO       | 

#### Path Parameters

| Name            | Description                                                                                           | Required | Default value |
| --------------- | ----------------------------------------------------------------------------------------------------- | -------- | ------------- |
| `{report_name}` | Name of the report that contains all the information about the profile, filter tags, group types etc. | YES      |


#### Notes
If user omits start_time and end_time parameters during the request the argo-web-api will respond with the latest status results of the current day

### Example Request 1: Status results (for today or for a specific period)

#### Request

##### Method
`HTTP GET`

##### Path

For today's results just issue:

```
/api/v2/status/Report_A
```

For results in a specific period:

```
/api/v2/status/Report_A?start_time=2015-06-20T12:00:00Z&end_time=2015-06-26T23:00:00Z 
```

##### Headers

```
x-api-key: "tenant_key_value"
Accept: "application/json"
```

#### Response

##### Code

```
Status: 200 OK
```

##### Body

```json
{
  "groups": [
    {
      "name": "SITEA",
      "type": "SITES",
      "statuses": [
        {
          "timestamp": "2015-05-01T00:00:00Z",
          "value": "OK"
        },
        {
          "timestamp": "2015-05-01T01:00:00Z",
          "value": "CRITICAL"
        },
        {
          "timestamp": "2015-05-01T05:00:00Z",
          "value": "OK"
        },
        {
          "timestamp": "2015-05-01T23:59:59Z",
          "value": "OK"
        }
      ],
      "endpoints": [
        {
          "hostname": "cream01.example.foo",
          "service": "CREAM-CE",
          "info": {
            "Url": "http://example.foo/path/to/service"
          },
          "statuses": [
            {
              "timestamp": "2015-05-01T00:00:00Z",
              "value": "OK"
            },
            {
              "timestamp": "2015-05-01T01:00:00Z",
              "value": "CRITICAL"
            },
            {
              "timestamp": "2015-05-01T05:00:00Z",
              "value": "OK"
            },
            {
              "timestamp": "2015-05-01T23:59:59Z",
              "value": "OK"
            }
          ]
        },
        {
          "hostname": "cream02.example.foo",
          "service": "CREAM-CE",
          "statuses": [
            {
              "timestamp": "2015-05-01T00:00:00Z",
              "value": "OK"
            },
            {
              "timestamp": "2015-05-01T08:47:00Z",
              "value": "WARNING"
            },
            {
              "timestamp": "2015-05-01T12:00:00Z",
              "value": "OK"
            },
            {
              "timestamp": "2015-05-01T23:59:59Z",
              "value": "OK"
            }
          ]
        }
      ]
    },
    {
      "name": "SITEB",
      "type": "SITES",
      "statuses": [
        {
          "timestamp": "2015-05-01T00:00:00Z",
          "value": "OK"
        },
        {
          "timestamp": "2015-05-01T03:00:00Z",
          "value": "WARNING"
        },
        {
          "timestamp": "2015-05-01T17:53:00Z",
          "value": "CRITICAL"
        },
        {
          "timestamp": "2015-05-01T23:59:59Z",
          "value": "CRITICAL"
        }
      ],
      "endpoints": [
        {
          "hostname": "cream03.example.foo",
          "service": "CREAM-CE",
          "statuses": [
            {
              "timestamp": "2015-05-01T00:00:00Z",
              "value": "OK"
            },
            {
              "timestamp": "2015-05-01T03:00:00Z",
              "value": "WARNING"
            },
            {
              "timestamp": "2015-05-01T17:53:00Z",
              "value": "CRITICAL"
            },
            {
              "timestamp": "2015-05-01T23:59:59Z",
              "value": "CRITICAL"
            }
          ]
        }
      ]
    }
  ]
}
```

### Example Request 2: Display only latest status results

#### Request

##### Method
`HTTP GET`

##### Path


```
/api/v2/status/Report_A?latest=true
```

##### Headers

```
x-api-key: "tenant_key_value"
Accept: "application/json"
```

#### Response

##### Code

```
Status: 200 OK
```

##### Body

```json
{
  "groups": [
    {
      "name": "SITEA",
      "type": "SITES",
      "statuses": [
        {
          "timestamp": "2022-05-11T15:00:00Z",
          "value": "OK"
        }
      ],
      "endpoints": [
        {
          "hostname": "cream01.example.foo",
          "service": "CREAM-CE",
          "info": {
            "Url": "http://example.foo/path/to/service"
          },
          "statuses": [
            {
              "timestamp": "2022-05-11T15:00:00Z",
              "value": "OK"
            }
          ]
        },
        {
          "hostname": "cream02.example.foo",
          "service": "CREAM-CE",
          "statuses": [
            {
              "timestamp": "2022-05-11T15:00:00Z",
              "value": "OK"
            }
          ]
        }
      ]
    },
    {
      "name": "SITEB",
      "type": "SITES",
      "statuses": [
        {
          "timestamp": "2022-05-11T15:00:00Z",
          "value": "CRITICAL"
        }
      ],
      "endpoints": [
        {
          "hostname": "cream03.example.foo",
          "service": "CREAM-CE",
          "statuses": [
            {
              "timestamp": "2022-05-11T15:00:00Z",
              "value": "CRITICAL"
            }
          ]
        }
      ]
    }
  ]
}
```
