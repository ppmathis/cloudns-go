# cloudns-go

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/ppmathis/cloudns-go/LICENSE.txt)
[![Documentation](http://img.shields.io/badge/docs-godoc.org-blue.svg)](https://godoc.org/github.com/ppmathis/cloudns-go)
[![Go Compatibility](https://img.shields.io/badge/golang-1.13+-brightgreen.svg)](#)
[![GitHub issues](https://img.shields.io/github/issues/ppmathis/cloudns-go.svg)](https://github.com/ppmathis/cloudns-go/issues)
[![Coverage Status](https://codecov.io/gh/ppmathis/cloudns-go/branch/master/graph/badge.svg?token=DMZR0O1H69)](https://codecov.io/gh/ppmathis/cloudns-go)
[![Copyright](https://img.shields.io/badge/copyright-Pascal_Mathis-lightgrey.svg)](#)

## Summary
This is an unofficial library for the ClouDNS HTTP API written in Go. Currently all operations related to account,
zone and record management have been fully implemented. Further information about the API can be found at the
[official ClouDNS website](https://www.cloudns.net/).

## Quick Start
Initialize cloudns-go by creating a new API client instance with your preferred choice of credentials, which is a
combination of the API user password and an user ID, sub-user ID or sub-user name:

```go
client, err := cloudns.New(
    // You must only specify one of these options
    // AuthUserID has the highest set of privileges and access to everything
    // AuthSubUserID and AuthSubUserName are restricted
    cloudns.AuthUserID(42, "cloudns-rocks"),
    cloudns.AuthSubUserID(13, "what-a-lucky-day"),
    cloudns.AuthSubUserName("john", "doe"),
)
```

After confirming that no error has occurred, you may access the various services available underneath the client object,
which currently consists of:

- `client.Accounts`: Manage your ClouDNS account and sub-users
- `client.Zones`: Manage DNS zones in your account
- `client.Records`: Manage records inside a specific DNS zone

You can find more information about the specific methods and structures of cloudns-go by visiting the
[official documentation on godoc.org](https://godoc.org/github.com/ppmathis/cloudns-go).


## Example
```go
package main

import (
	"context"
	"fmt"
	"github.com/ppmathis/cloudns-go"
)

func main() {
	client, _ := cloudns.New(
		cloudns.AuthUserID(42, "cloudns-rocks"),
	)

	zone, _ := client.Zones.Get(context.TODO(), "api-example.com")
	result1, _ := client.Zones.SetActive(context.TODO(), zone.Name, true)

	record := cloudns.NewRecord(cloudns.RecordTypeA, "localhost", "1.2.3.4", 3600)
	result2, _ := client.Records.Create(context.TODO(), zone.Name, record)

	fmt.Printf("Zone: %+v\n", zone)
	fmt.Printf("Record: %+v\n", record)
	fmt.Printf("Result of `Zones.SetActive()`: %+v\n", result1)
	fmt.Printf("Result of `Records.Create()`: %+v\n", result2)
}
```
