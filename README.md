# cloudns-go

[![License](https://img.shields.io/badge/license-GPL--3.0+-blue.svg)](https://github.com/snapserv/cloudns-go/LICENSE.txt)
[![Documentation](http://img.shields.io/badge/docs-godoc.org-blue.svg)](https://godoc.org/github.com/snapserv/cloudns-go)
[![Go Compatibility](https://img.shields.io/badge/golang-1.13-brightgreen.svg)](#)
[![GitHub issues](https://img.shields.io/github/issues/snapserv/cloudns-go.svg)](https://github.com/snapserv/cloudns-go/issues)
[![Coverage Status](https://coveralls.io/repos/github/snapserv/cloudns-go/badge.svg?branch=master)](https://coveralls.io/github/snapserv/cloudns-go?branch=master)
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
[official documentation on godoc.org](https://godoc.org/github.com/snapserv/cloudns-go). 
